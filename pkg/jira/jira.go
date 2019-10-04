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
	issues, err := api.Issues(jql, 0)

	if err != nil {
		log.Println("Could not get issues.", err)
		return pkg.Timesheet{}
	}

	burstLimit := 5
	// If we do not throttle here, the server will sometimes respond with 401.
	// That looks like a rate limit but I have not found any numbers in the docs for the cloud setup.
	throttle := make(chan struct{}, burstLimit)
	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(item model.Issue) {
			defer wg.Done()
			throttle <- struct{}{}
			items, err := api.Worklog(item.Key(), 0)
			<-throttle
			if err != nil {
				log.Println("Could not get effort for "+item.Key(), err)
				return
			}
			worklog := item.Worklog(accounts, items, fromDate, toDate)
			for account := range accounts {
				c <- worklog[account]
			}
		}(item)
	}
	go func() {
		defer close(c)
		defer close(throttle)
		wg.Wait()
	}()
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort...)
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
