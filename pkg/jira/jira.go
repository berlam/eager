package jira

import (
	"bytes"
	"eager/pkg"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	headerAccountId      = "X-AACCOUNTID"
	jiraSearchProjectUrl = "/rest/api/3/project/search?startAt=%s"
	jiraSearchUserUrl    = "/rest/api/3/user/assignable/multiProjectSearch?projectKeys=%s&query=%s&maxResults=2"
	jiraSearchIssueUrl   = "/rest/api/3/search"
	jiraWorklogUrl       = "/rest/api/3/issue/%s/worklog?startAt=%s"
)

func GetTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []string) pkg.Timesheet {
	fromDate, toDate := pkg.GetTimeRange(year, month)

	jql := new(jql).between(fromDate, toDate).me().projects(projects...)
	accountId, issues, err := getIssues(client, server, userinfo, jql, 0)

	if err != nil {
		log.Println("Could not get Jira issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(wg *sync.WaitGroup, item issue) {
			defer wg.Done()
			c <- item.getWorklog(client, server, userinfo, fromDate, toDate)[accountId]
		}(&wg, item)
	}
	go func(wg *sync.WaitGroup) {
		defer close(c)
		wg.Wait()
	}(&wg)
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort...)
	}

	return timesheet
}

func GetBulkTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects, users []string) pkg.Timesheet {
	var err error
	fromDate, toDate := pkg.GetTimeRange(year, month)

	if projects == nil {
		projects, err = getProjects(client, server, userinfo, 0)
		if err != nil {
			log.Println("Could not get projects.", err)
			return pkg.Timesheet{}
		}
	}

	if users != nil {
		accounts, err := getAccounts(client, server, userinfo, projects, users)
		if err != nil {
			log.Println("Could not get user.", err)
			return pkg.Timesheet{}
		}
		i := 0
		for _, account := range accounts {
			users[i] = string(account)
			i++
		}
	}

	jql := new(jql).between(fromDate, toDate).users(users...).projects(projects...)
	_, issues, err := getIssues(client, server, userinfo, jql, 0)

	if err != nil {
		log.Println("Could not get issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Timesheet)
	var wg sync.WaitGroup
	wg.Add(len(issues))
	for _, item := range issues {
		go func(wg *sync.WaitGroup, item issue) {
			defer wg.Done()
			items := item.getWorklog(client, server, userinfo, fromDate, toDate)
			for _, user := range users {
				c <- items[accountId(user)]
			}
		}(&wg, item)
	}
	go func(wg *sync.WaitGroup) {
		defer close(c)
		wg.Wait()
	}(&wg)
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

func getProjects(client *http.Client, server *url.URL, userinfo *url.Userinfo, startAt int) ([]string, error) {
	projectUrl, err := server.Parse(fmt.Sprintf(jiraSearchProjectUrl, strconv.Itoa(startAt)))
	response, err := createRequest(client, http.MethodGet, projectUrl, userinfo, nil)
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

	var result = projectQueryResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	projectKeys := make([]string, 0, result.Total)
	for _, project := range result.Values {
		projectKeys = append(projectKeys, string(project.ProjectKey))
	}
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		nextProjectKeys, err := getProjects(client, server, userinfo, result.Total)
		if err != nil {
			return nil, err
		}
		projectKeys = append(projectKeys, nextProjectKeys...)
	}
	return projectKeys, nil
}

func getAccounts(client *http.Client, server *url.URL, userinfo *url.Userinfo, projects, users []string) (map[string]accountId, error) {
	result := make(map[string]accountId, len(users))
	c := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(users))
	for _, user := range users {
		go func(wg *sync.WaitGroup, user string) {
			defer wg.Done()
			account, err := getUsersByProjects(client, server, userinfo, user, projects)
			if err != nil {
				c <- err
				return
			}
			result[user] = account
		}(&wg, user)
	}
	go func(wg *sync.WaitGroup) {
		defer close(c)
		wg.Wait()
	}(&wg)
	for err := range c {
		return nil, err
	}
	return result, nil
}

func getUsersByProjects(client *http.Client, server *url.URL, userinfo *url.Userinfo, user string, projects []string) (accountId, error) {
	userUrl, err := server.Parse(fmt.Sprintf(jiraSearchUserUrl, strings.Join(projects, ","), user))
	response, err := createRequest(client, http.MethodGet, userUrl, userinfo, nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf(response.Status)
	}

	var result = make([]userQueryResult, 0, 2)
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", fmt.Errorf("found no user for %s", user)
	}
	if len(result) > 1 {
		return "", fmt.Errorf("found more than one user for %s", user)
	}
	return result[0].AccountId, nil
}

func getIssues(client *http.Client, server *url.URL, userinfo *url.Userinfo, jql jql, startAt int) (accountId, []issue, error) {
	body, _ := json.Marshal(issueQuery{
		Fields:         []string{"project"},
		Jql:            jql.build(),
		PaginatedQuery: &PaginatedQuery{StartAt: startAt},
	})
	searchUrl, _ := server.Parse(jiraSearchIssueUrl)
	response, err := createRequest(client, http.MethodPost, searchUrl, userinfo, bytes.NewBuffer(body))
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	account, _ := url.PathUnescape(response.Header.Get(headerAccountId))
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil, err
	}
	if response.StatusCode != 200 {
		return "", nil, fmt.Errorf(response.Status)
	}

	var result = issueQueryResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", nil, err
	}
	issues := result.Issues
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		_, nextIssues, err := getIssues(client, server, userinfo, jql, result.Total)
		if err != nil {
			return "", nil, err
		}
		issues = append(issues, nextIssues...)
	}
	return accountId(account), issues, err
}

func getWorklogItems(client *http.Client, server *url.URL, userinfo *url.Userinfo, key issueKey, startAt int) (worklogItems, error) {
	worklogUrl, err := server.Parse(fmt.Sprintf(jiraWorklogUrl, string(key), strconv.Itoa(startAt)))
	response, err := createRequest(client, http.MethodGet, worklogUrl, userinfo, nil)
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

	var result = worklogQueryResult{}
	err = json.Unmarshal(data, &result)
	items := result.Worklogs
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		nextItems, err := getWorklogItems(client, server, userinfo, key, result.Total)
		if err != nil {
			return nil, err
		}
		items = append(items, nextItems...)
	}
	return items, nil
}

func (items worklogItems) getWorklogByAccountIdAndDate(pKey projectKey, iKey issueKey, fromDate, toDate time.Time) map[accountId]pkg.Timesheet {
	result := make(map[accountId]pkg.Timesheet)
	total := len(items)
	for _, effort := range items {
		if effort.isBetween(fromDate, toDate) {
			current := result[effort.Author.AccountId]
			if current == nil {
				current = make(pkg.Timesheet, 0, total)
			}
			result[effort.Author.AccountId] = append(current, pkg.Effort{
				Employee:    pkg.Employee(effort.Author.DisplayName),
				Description: effort.getComment(),
				Project:     pkg.Project(pKey),
				Task:        pkg.Task(iKey),
				Date:        effort.getDate(),
				Duration:    effort.getDuration(),
			})
		}
	}
	return result
}

func (issue issue) getWorklog(client *http.Client, server *url.URL, userinfo *url.Userinfo, fromDate, toDate time.Time) map[accountId]pkg.Timesheet {
	worklogItems, err := getWorklogItems(client, server, userinfo, issue.Key, 0)
	if err != nil {
		log.Println("Could not get effort for "+issue.Key, err)
		return map[accountId]pkg.Timesheet{}
	}
	return worklogItems.getWorklogByAccountIdAndDate(issue.Fields.Project.Key, issue.Key, fromDate, toDate)
}

func (effort worklogItem) getDate() time.Time {
	date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
	return date.UTC().Truncate(time.Hour * 24)
}

func (effort worklogItem) isBetween(fromDate, toDate time.Time) bool {
	date := effort.getDate()
	return !date.Before(fromDate) && date.Before(toDate)
}

func (effort worklogItem) getComment() pkg.Description {
	description := ""
	if len(effort.Comment.Content) > 0 && len(effort.Comment.Content[0].Content) > 0 {
		description = effort.Comment.Content[0].Content[0].Text
	}
	return pkg.Description(description)
}

func (effort worklogItem) getDuration() time.Duration {
	duration, _ := time.ParseDuration(strconv.Itoa(effort.TimeSpentSeconds) + "s")
	return duration
}

type accountId string
type projectKey string
type issueKey string

type author struct {
	AccountId    accountId `json:"accountId"`
	EmailAddress string    `json:"emailAddress"`
	DisplayName  string    `json:"displayName"`
}

type PaginatedQuery struct {
	StartAt int `json:"startAt"`
}

type PaginatedResult struct {
	MaxResults int   `json:"maxResults"`
	StartAt    int   `json:"startAt"`
	Total      int   `json:"total"`
	IsLast     *bool `json:"isLast,omitempty"`
}

type projectQueryResult struct {
	*PaginatedResult
	Values []struct {
		ProjectKey projectKey `json:"key"`
	} `json:"values"`
}

type userQueryResult struct {
	AccountId accountId `json:"accountId"`
}

type issueQuery struct {
	*PaginatedQuery
	Fields []string `json:"fields"`
	Jql    string   `json:"jql"`
}

type issueQueryResult struct {
	*PaginatedResult
	Issues []issue `json:"issues"`
}

type issue struct {
	Id     string   `json:"id"`
	Key    issueKey `json:"key"`
	Fields struct {
		Project struct {
			Id   string     `json:"id"`
			Key  projectKey `json:"key"`
			Name string     `json:"name"`
		} `json:"project"`
	} `json:"fields"`
}

type worklogQueryResult struct {
	*PaginatedResult
	Worklogs worklogItems `json:"worklogs"`
}

type worklogItems []*worklogItem

type worklogItem struct {
	Author       author `json:"author"`
	UpdateAuthor author `json:"updateAuthor"`
	Comment      struct {
		Type    string `json:"type"`
		Version int    `json:"version"`
		Content []*struct {
			Type    string `json:"type"`
			Content []*struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	} `json:"comment,omitempty"`
	Started          string `json:"started"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}
