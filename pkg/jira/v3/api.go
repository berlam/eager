package v3

import (
	"bytes"
	"eager/pkg"
	"eager/pkg/jira/model"
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
	searchProjectUrl = "/rest/api/3/project/search?startAt=%s"
	searchUserUrl    = "/rest/api/3/user/assignable/multiProjectSearch?projectKeys=%s&query=%s&maxResults=2"
	searchIssueUrl   = "/rest/api/3/search"
	worklogUrl       = "/rest/api/3/issue/%s/worklog?startAt=%s"
)

type Api struct {
	Client   *http.Client
	Server   *url.URL
	Userinfo *url.Userinfo
}

func (api Api) Projects(startAt int) ([]pkg.Project, error) {
	projectUrl, err := api.Server.Parse(fmt.Sprintf(searchProjectUrl, strconv.Itoa(startAt)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, projectUrl, api.Userinfo, nil)
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
		nextProjectKeys, err := api.Projects(result.Total)
		if err != nil {
			return nil, err
		}
		projectKeys = append(projectKeys, nextProjectKeys...)
	}
	return projectKeys, nil
}

func (api Api) Accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]model.Account, error) {
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

func (api Api) User(user pkg.User, projects []pkg.Project) (model.Account, error) {
	projectQueryPart := make([]string, len(projects))
	for i, project := range projects {
		projectQueryPart[i] = string(project)
	}
	userUrl, err := api.Server.Parse(fmt.Sprintf(searchUserUrl, strings.Join(projectQueryPart, ","), user))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, userUrl, api.Userinfo, nil)
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

func (api Api) Issues(jql model.Jql, startAt int) (model.Account, []model.Issue, error) {
	body, _ := json.Marshal(issueQuery{
		Fields:         []string{"project"},
		Jql:            jql.Build(),
		PaginatedQuery: &PaginatedQuery{StartAt: startAt},
	})
	searchUrl, _ := api.Server.Parse(searchIssueUrl)
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodPost, searchUrl, api.Userinfo, bytes.NewBuffer(body))
	if err != nil {
		return "", nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	account, _ := url.PathUnescape(response.Header.Get(model.HeaderAccountId))
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
		_, nextIssues, err := api.Issues(jql, result.Total)
		if err != nil {
			return "", nil, err
		}
		issues = append(issues, nextIssues...)
	}
	return model.Account(account), issues, err
}

func (api Api) Worklog(key model.IssueKey, startAt int) ([]model.Worklog, error) {
	worklogUrl, err := api.Server.Parse(fmt.Sprintf(worklogUrl, string(key), strconv.Itoa(startAt)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, worklogUrl, api.Userinfo, nil)
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
		nextItems, err := api.Worklog(key, result.Total)
		if err != nil {
			return nil, err
		}
		items = append(items, nextItems...)
	}
	return items, nil
}

func (result issueQueryResult) issues() []model.Issue {
	issues := make([]model.Issue, len(result.ApiIssues))
	for idx, issue := range result.ApiIssues {
		issues[idx] = issue
	}
	return issues
}

func (result worklogQueryResult) worklogs() []model.Worklog {
	worklogs := make([]model.Worklog, len(result.ApiWorklogs))
	for idx, worklog := range result.ApiWorklogs {
		worklogs[idx] = worklog
	}
	return worklogs
}

func (issue issue) Key() model.IssueKey {
	return issue.ApiKey
}

func (issue issue) Worklog(worklog []model.Worklog, fromDate, toDate time.Time) map[model.Account]pkg.Timesheet {
	pKey := issue.Fields.Project.Key
	iKey := issue.ApiKey
	result := make(map[model.Account]pkg.Timesheet)
	total := len(worklog)
	for _, effort := range worklog {
		if effort.IsBetween(fromDate, toDate) {
			current := result[effort.Author().Id()]
			if current == nil {
				current = make(pkg.Timesheet, 0, total)
			}
			result[effort.Author().Id()] = append(current, pkg.Effort{
				User:        pkg.User(effort.Author().Name()),
				Description: effort.Comment(),
				Project:     pkg.Project(pKey),
				Task:        pkg.Task(iKey),
				Date:        effort.Date(),
				Duration:    effort.Duration(),
			})
		}
	}
	return result
}

func (author author) Id() model.Account {
	return author.AccountId
}

func (author author) Name() string {
	return author.DisplayName
}

func (effort worklogItem) Author() model.Author {
	return effort.ApiAuthor
}

func (effort worklogItem) IsBetween(fromDate, toDate time.Time) bool {
	date := effort.Date()
	return !date.Before(fromDate) && date.Before(toDate)
}

func (effort worklogItem) Date() time.Time {
	date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
	return date.UTC().Truncate(time.Hour * 24)
}

func (effort worklogItem) Comment() pkg.Description {
	description := ""
	if len(effort.ApiComment.Content) > 0 && len(effort.ApiComment.Content[0].Content) > 0 {
		description = effort.ApiComment.Content[0].Content[0].Text
	}
	return pkg.Description(description)
}

func (effort worklogItem) Duration() time.Duration {
	duration, _ := time.ParseDuration(strconv.Itoa(effort.TimeSpentSeconds) + "s")
	return duration
}
