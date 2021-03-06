package jira

import (
	"eager/pkg"
	"eager/pkg/jira/cloud"
	"eager/pkg/jira/model"
	"eager/pkg/jira/v2"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	jiraServerInfo = "/rest/api/latest/serverInfo"
)

func getApiVersion(client *http.Client, server *url.URL, userinfo *url.Userinfo) (model.Api, error) {
	infoUrl, err := server.Parse(fmt.Sprintf(jiraServerInfo))
	response, err := pkg.CreateJsonRequest(client, http.MethodGet, infoUrl, userinfo, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf(response.Status)
	}

	var result = serverInfo{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	// There are "Cloud" and "Server" deployment types.
	if strings.ToLower(result.DeploymentType) == "cloud" {
		path, _ := server.Parse(cloud.BasePath)
		return &cloud.Api{
			Client:   client,
			Server:   path,
			Userinfo: userinfo,
		}, nil
	}
	path, _ := server.Parse(v2.BasePath)
	return &v2.Api{
		Client:   client,
		Server:   path,
		Userinfo: userinfo,
	}, nil
}

func GetTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []pkg.Project) pkg.Timesheet {
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return pkg.Timesheet{}
	}

	accountId, timezone, err := api.Me()
	if err != nil {
		log.Println("Could not get user.", err)
		return pkg.Timesheet{}
	}
	accounts := map[model.Account]*pkg.User{}
	accounts[accountId] = &pkg.User{
		TimeZone: timezone,
	}

	return do(api, year, month, projects, accounts)
}

func GetBulkTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []pkg.Project, users []*pkg.User) pkg.Timesheet {
	var err error
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return pkg.Timesheet{}
	}

	accounts, err := accounts(api, users)
	if err != nil {
		log.Println("Could not get user.", err)
		return pkg.Timesheet{}
	}

	return do(api, year, month, projects, accounts)
}

func AddWorklogItem(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, day int, task pkg.Task, duration time.Duration, sum bool, confirm model.WorklogFunc) {
	var err error
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return
	}

	account, location, err := api.Me()
	if err != nil {
		log.Println("Could not get user.", err)
		return
	}

	key := model.IssueKey(task)

	if !sum {
		// Add new effort
		err = api.AddWorklog(key, adjustDateTime(location, duration, year, month, day), duration)
		if err != nil {
			log.Println("Could not add effort.", err)
		}
		return
	}

	// Check, if there is already effort inside the worklog
	var effort []model.Worklog
	err = api.Worklog(key, func(worklog model.Worklog) bool {
		wd := worklog.Date().In(location)
		if worklog.Author().Id() == account && year == wd.Year() && month == wd.Month() && day == wd.Day() {
			effort = append(effort, worklog)
		}
		return true
	})
	if err != nil {
		log.Println("Could not get worklog.", err)
		return
	}

	// Collect effort for that day
	for _, worklog := range effort {
		duration += worklog.Duration()
	}

	// Add new effort
	err = api.AddWorklog(key, adjustDateTime(location, duration, year, month, day), duration)
	if err != nil {
		log.Println("Could not add effort.", err)
		return
	}

	// Delete old effort
	for _, worklog := range effort {
		if confirm(worklog) {
			err = api.RemoveWorklog(key, worklog.Id())
			if err != nil {
				log.Println("Could not remove effort.", err)
			}
		}
	}
}

func adjustDateTime(location *time.Location, duration time.Duration, year int, month time.Month, day int) time.Time {
	// Get the current date and time
	// Sub the given duration
	date := time.Now().In(location).Add(-duration)
	// If we have a different day then, adjust the time to 00:00:00
	// Also applies, if we are not on the current day
	if date.Year() != year || date.Month() != month || date.Day() != day {
		date = time.Date(year, month, day, 0, 0, 0, 0, location)
	}
	return date
}

func RemoveWorklogItem(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, day int, task pkg.Task, confirm model.WorklogFunc) {
	var err error
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return
	}

	account, location, err := api.Me()
	if err != nil {
		log.Println("Could not get user.", err)
		return
	}

	key := model.IssueKey(task)
	date := time.Date(year, month, day, 0, 0, 0, 0, location)

	// Check, if there is already effort inside the worklog
	err = api.Worklog(key, func(worklog model.Worklog) bool {
		wd := worklog.Date()
		if worklog.Author().Id() == account && date.Year() == wd.Year() && date.Month() == wd.Month() && date.Day() == wd.Day() {
			if confirm(worklog) {
				err = api.RemoveWorklog(key, worklog.Id())
				if err != nil {
					log.Println("Could not remove effort.", err)
					return false
				}
			}
		}
		return true
	})
}

func do(api model.Api, year int, month time.Month, projects []pkg.Project, accounts map[model.Account]*pkg.User) pkg.Timesheet {
	var err error

	// TODO Calculate max timezone offset for each user to have the right from and to date.
	// The jql query uses afaik the time zone of the requesting user.
	fromDate, toDate := pkg.GetTimeRange(year, month)

	i := 0
	accountIds := make([]model.Account, len(accounts))
	for account := range accounts {
		accountIds[i] = account
		i++
	}

	jql := new(model.Jql).Between(fromDate, toDate).Users(accountIds...).Projects(projects...)

	// The chan for errors
	errors := make(chan error)

	// The chan for issues
	issues := make(chan model.Issue)
	go func() {
		defer close(issues)
		err = api.Issues(jql, func(issue model.Issue) {
			issues <- issue
		})
		if err != nil {
			log.Println("Could not get issues.", err)
			errors <- err
		}
	}()

	// The chan for effort
	effort := make(chan pkg.Effort)
	go func() {
		var wg sync.WaitGroup
		throttle := make(chan struct{}, 5)
		defer close(effort)
		defer close(throttle)
		for issue := range issues {
			wg.Add(1)
			throttle <- struct{}{}
			go func(issue model.Issue) {
				defer func() {
					<-throttle
					wg.Done()
				}()
				err = api.Worklog(issue.Key(), func(worklog model.Worklog) bool {
					account := worklog.Author().Id()
					user := accounts[account]
					if user == nil {
						return true
					}
					date := worklog.Date().In(user.TimeZone)
					date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
					if !date.Before(fromDate) && date.Before(toDate) {
						effort <- pkg.Effort{
							User:        user,
							Description: worklog.Comment(),
							Project:     issue.Project(),
							Task:        pkg.Task(issue.Key()),
							Date:        date,
							Duration:    worklog.Duration(),
						}
					}
					return true
				})
				if err != nil {
					log.Println("Could not get effort for "+issue.Key(), err)
					errors <- err
				}
			}(issue)
		}
		wg.Wait()
	}()

	var timesheet pkg.Timesheet
	go func() {
		defer close(errors)
		for e := range effort {
			timesheet = append(timesheet, e)
		}
	}()
	if <-errors != nil {
		return pkg.Timesheet{}
	}

	return timesheet
}

func accounts(api model.Api, users []*pkg.User) (map[model.Account]*pkg.User, error) {
	result := make(map[model.Account]*pkg.User, len(users))
	c := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(users))
	for _, user := range users {
		go func(user *pkg.User) {
			defer wg.Done()
			account, location, err := api.User(user)
			if err != nil {
				c <- err
				return
			}
			result[account] = &pkg.User{
				DisplayName: user.DisplayName,
				TimeZone:    location,
			}
		}(user)
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()
	for err := range c {
		return nil, err
	}
	return result, nil
}

type serverInfo struct {
	VersionNumbers []int  `json:"versionNumbers"`
	DeploymentType string `json:"deploymentType"`
	ServerTitle    string `json:"serverTitle"`
}
