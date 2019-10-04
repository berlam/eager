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
	fromDate, toDate := pkg.GetTimeRange(year, month)

	jql := new(model.Jql).Between(fromDate, toDate).Me().Projects(projects...)
	accountId, issues, err := api.Issues(jql, 0)

	if err != nil {
		log.Println("Could not get Jira issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(item model.Issue) {
			defer wg.Done()
			items, err := api.Worklog(item.Key(), 0)
			if err != nil {
				log.Println("Could not get effort for "+item.Key(), err)
				return
			}
			c <- item.Worklog(items, fromDate, toDate)[accountId]
		}(item)
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort...)
	}

	return timesheet
}

func GetBulkTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []pkg.Project, users []pkg.User) pkg.Timesheet {
	var err error
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return pkg.Timesheet{}
	}
	fromDate, toDate := pkg.GetTimeRange(year, month)

	if projects == nil || len(projects) == 0 {
		projects, err = api.Projects(0)
		if err != nil {
			log.Println("Could not get projects.", err)
			return pkg.Timesheet{}
		}
	}

	if users != nil && len(users) > 0 {
		accounts, err := accounts(api, projects, users)
		if err != nil {
			log.Println("Could not get user.", err)
			return pkg.Timesheet{}
		}
		i := 0
		for _, account := range accounts {
			users[i] = pkg.User(account)
			i++
		}
	}

	jql := new(model.Jql).Between(fromDate, toDate).Users(users...).Projects(projects...)
	_, issues, err := api.Issues(jql, 0)

	if err != nil {
		log.Println("Could not get issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(item model.Issue) {
			defer wg.Done()
			items, err := api.Worklog(item.Key(), 0)
			if err != nil {
				log.Println("Could not get effort for "+item.Key(), err)
				return
			}
			worklog := item.Worklog(items, fromDate, toDate)
			for _, user := range users {
				c <- worklog[model.Account(user)]
			}
		}(item)
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort...)
	}

	return timesheet
}

func accounts(api model.Api, projects []pkg.Project, users []pkg.User) (map[pkg.User]model.Account, error) {
	result := make(map[pkg.User]model.Account, len(users))
	c := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(users))
	for _, user := range users {
		go func(user pkg.User) {
			defer wg.Done()
			account, err := api.User(user, projects)
			if err != nil {
				c <- err
				return
			}
			result[user] = account
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
