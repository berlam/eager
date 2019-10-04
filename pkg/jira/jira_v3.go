package jira

import (
	"bytes"
	"eager/pkg"
	"encoding/json"
	"fmt"
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
	jiraSearchProjectUrl = "/rest/api/3/project/search?startAt=%s"
	jiraSearchUserUrl    = "/rest/api/3/user/assignable/multiProjectSearch?projectKeys=%s&query=%s&maxResults=2"
	jiraSearchIssueUrl   = "/rest/api/3/search"
	jiraWorklogUrl       = "/rest/api/3/issue/%s/worklog?startAt=%s"
)

type v3 struct {
	client   *http.Client
	server   *url.URL
	userinfo *url.Userinfo
}

func (api v3) projects(startAt int) ([]pkg.Project, error) {
	projectUrl, err := api.server.Parse(fmt.Sprintf(jiraSearchProjectUrl, strconv.Itoa(startAt)))
	response, err := createRequest(api.client, http.MethodGet, projectUrl, api.userinfo, nil)
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

	var result = projectQueryResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	projectKeys := make([]pkg.Project, 0, result.Total)
	for _, project := range result.Values {
		projectKeys = append(projectKeys, pkg.Project(project.ProjectKey))
	}
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		nextProjectKeys, err := api.projects(result.Total)
		if err != nil {
			return nil, err
		}
		projectKeys = append(projectKeys, nextProjectKeys...)
	}
	return projectKeys, nil
}

func (api v3) accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]Account, error) {
	result := make(map[pkg.User]Account, len(users))
	c := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(users))
	for _, user := range users {
		go func(user pkg.User) {
			defer wg.Done()
			account, err := api.user(user, projects)
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

func (api v3) user(user pkg.User, projects []pkg.Project) (Account, error) {
	projectQueryPart := make([]string, len(projects))
	for i, project := range projects {
		projectQueryPart[i] = string(project)
	}
	userUrl, err := api.server.Parse(fmt.Sprintf(jiraSearchUserUrl, strings.Join(projectQueryPart, ","), user))
	response, err := createRequest(api.client, http.MethodGet, userUrl, api.userinfo, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

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

func (api v3) issues(jql jql, startAt int) (Account, []Issue, error) {
	body, _ := json.Marshal(issueQuery{
		Fields:         []string{"project"},
		Jql:            jql.build(),
		PaginatedQuery: &PaginatedQuery{StartAt: startAt},
	})
	searchUrl, _ := api.server.Parse(jiraSearchIssueUrl)
	response, err := createRequest(api.client, http.MethodPost, searchUrl, api.userinfo, bytes.NewBuffer(body))
	if err != nil {
		return "", nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

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
	issues := result.issues()
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		_, nextIssues, err := api.issues(jql, result.Total)
		if err != nil {
			return "", nil, err
		}
		issues = append(issues, nextIssues...)
	}
	return Account(account), issues, err
}

func (api v3) worklog(key IssueKey, startAt int) ([]Worklog, error) {
	worklogUrl, err := api.server.Parse(fmt.Sprintf(jiraWorklogUrl, string(key), strconv.Itoa(startAt)))
	response, err := createRequest(api.client, http.MethodGet, worklogUrl, api.userinfo, nil)
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

	var result = worklogQueryResult{}
	err = json.Unmarshal(data, &result)
	items := result.worklogs()
	if (result.IsLast == nil && result.Total >= result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		nextItems, err := api.worklog(key, result.Total)
		if err != nil {
			return nil, err
		}
		items = append(items, nextItems...)
	}
	return items, nil
}

type projectKey string

type serverInfo struct {
	VersionNumbers []int  `json:"versionNumbers"`
	DeploymentType string `json:"deploymentType"`
	ServerTitle    string `json:"serverTitle"`
}

type author struct {
	AccountId    Account `json:"accountId"`
	EmailAddress string  `json:"emailAddress"`
	DisplayName  string  `json:"displayName"`
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
	AccountId Account `json:"accountId"`
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

func (result issueQueryResult) issues() []Issue {
	issues := make([]Issue, len(result.Issues))
	for e := range issues {
		issues[e] = result.Issues[e]
	}
	return issues
}

type issue struct {
	Id     string   `json:"id"`
	Key    IssueKey `json:"key"`
	Fields struct {
		Project struct {
			Id   string     `json:"id"`
			Key  projectKey `json:"key"`
			Name string     `json:"name"`
		} `json:"project"`
	} `json:"fields"`
}

func (issue issue) key() IssueKey {
	return issue.Key
}

func (issue issue) worklog(worklog []Worklog, fromDate, toDate time.Time) map[Account]pkg.Timesheet {
	pKey := issue.Fields.Project.Key
	iKey := issue.Key
	result := make(map[Account]pkg.Timesheet)
	total := len(worklog)
	for _, effort := range worklog {
		if effort.isBetween(fromDate, toDate) {
			current := result[effort.author().id()]
			if current == nil {
				current = make(pkg.Timesheet, 0, total)
			}
			result[effort.author().id()] = append(current, pkg.Effort{
				User:        pkg.User(effort.author().name()),
				Description: effort.comment(),
				Project:     pkg.Project(pKey),
				Task:        pkg.Task(iKey),
				Date:        effort.date(),
				Duration:    effort.duration(),
			})
		}
	}
	return result
}

type worklogQueryResult struct {
	*PaginatedResult
	Worklogs []*worklogItem `json:"worklogs"`
}

func (result worklogQueryResult) worklogs() []Worklog {
	worklogs := make([]Worklog, len(result.Worklogs))
	for e := range worklogs {
		worklogs[e] = result.Worklogs[e]
	}
	return worklogs
}

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

func (author author) id() Account {
	return author.AccountId
}

func (author author) name() string {
	return author.DisplayName
}

func (effort worklogItem) author() Author {
	return effort.Author
}

func (effort worklogItem) isBetween(fromDate, toDate time.Time) bool {
	date := effort.date()
	return !date.Before(fromDate) && date.Before(toDate)
}

func (effort worklogItem) date() time.Time {
	date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
	return date.UTC().Truncate(time.Hour * 24)
}

func (effort worklogItem) comment() pkg.Description {
	description := ""
	if len(effort.Comment.Content) > 0 && len(effort.Comment.Content[0].Content) > 0 {
		description = effort.Comment.Content[0].Content[0].Text
	}
	return pkg.Description(description)
}

func (effort worklogItem) duration() time.Duration {
	duration, _ := time.ParseDuration(strconv.Itoa(effort.TimeSpentSeconds) + "s")
	return duration
}
