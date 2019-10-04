package jira

import (
	"eager/pkg"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	headerAccountId      = "X-AACCOUNTID"
	jiraServerInfo       = "/rest/api/latest/serverInfo"
	jiraSearchProjectUrl = "/rest/api/3/project/search?startAt=%s"
	jiraSearchUserUrl    = "/rest/api/3/user/assignable/multiProjectSearch?projectKeys=%s&query=%s&maxResults=2"
	jiraSearchIssueUrl   = "/rest/api/3/search"
	jiraWorklogUrl       = "/rest/api/3/issue/%s/worklog?startAt=%s"
)

func getApiVersion(client *http.Client, server *url.URL, userinfo *url.Userinfo) (Api, error) {
	infoUrl, err := server.Parse(fmt.Sprintf(jiraServerInfo))
	response, err := createRequest(client, http.MethodGet, infoUrl, userinfo, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

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
		return &v3{
			client:   client,
			server:   server,
			userinfo: userinfo,
		}, nil
	}
	return nil, nil
}

func GetTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []pkg.Project) pkg.Timesheet {
	api, err := getApiVersion(client, server, userinfo)
	if err != nil {
		log.Println("Could not get api version.", err)
		return pkg.Timesheet{}
	}
	fromDate, toDate := pkg.GetTimeRange(year, month)

	jql := new(jql).between(fromDate, toDate).me().projects(projects...)
	accountId, issues, err := api.issues(jql, 0)

	if err != nil {
		log.Println("Could not get Jira issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(item issue) {
			defer wg.Done()
			items, err := api.worklog(item.Key, 0)
			if err != nil {
				log.Println("Could not get effort for "+item.Key, err)
				return
			}
			c <- item.getWorklog(items, fromDate, toDate)[accountId]
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
		projects, err = api.projects(0)
		if err != nil {
			log.Println("Could not get projects.", err)
			return pkg.Timesheet{}
		}
	}

	if users != nil && len(users) > 0 {
		accounts, err := api.accounts(projects, users)
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

	jql := new(jql).between(fromDate, toDate).users(users...).projects(projects...)
	_, issues, err := api.issues(jql, 0)

	if err != nil {
		log.Println("Could not get issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(item issue) {
			defer wg.Done()
			items, err := api.worklog(item.Key, 0)
			if err != nil {
				log.Println("Could not get effort for "+item.Key, err)
				return
			}
			worklog := item.getWorklog(items, fromDate, toDate)
			for _, user := range users {
				c <- worklog[Account(user)]
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

func createRequest(client *http.Client, httpMethod string, server *url.URL, userinfo *url.Userinfo, payload io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(httpMethod, server.String(), payload)
	if err != nil {
		return nil, err
	}
	password, _ := userinfo.Password()
	request.SetBasicAuth(userinfo.Username(), password)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, err
}

type Account string

type IssueKey string

type Api interface {
	ProjectAccessor
	AccountAccessor
	UserAccessor
	IssueAccessor
	WorklogAccessor
}

type ProjectAccessor interface {
	projects(startAt int) ([]pkg.Project, error)
}

type AccountAccessor interface {
	accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]Account, error)
}

type UserAccessor interface {
	user(user pkg.User, projects []pkg.Project) (Account, error)
}

type IssueAccessor interface {
	issues(jql jql, startAt int) (Account, []issue, error)
}

type WorklogAccessor interface {
	worklog(key IssueKey, startAt int) (worklogItems, error)
}
