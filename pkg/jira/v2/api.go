package v2

import (
	"bytes"
	"eager/pkg"
	"eager/pkg/jira/model"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	BasePath         = "/rest/api/2/"
	myselfUrl        = "myself"
	getUserUrl       = "user?accountId=%s"
	searchUserUrl    = "user/search?query=%s&maxResults=2"
	searchIssueUrl   = "search"
	getWorklogUrl    = "issue/%s/worklog?startAt=%s"
	addWorklogUrl    = "issue/%s/worklog?notifyUsers=false&adjustEstimate=leave"
	removeWorklogUrl = "issue/%s/worklog/%s?notifyUsers=false&adjustEstimate=leave"
)

type Api struct {
	Client   *http.Client
	Server   *url.URL
	Userinfo *url.Userinfo
}

func (api Api) Me() (model.Account, *time.Location, error) {
	myselfUrl, err := api.Server.Parse(myselfUrl)
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, myselfUrl, api.Userinfo, nil)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	reader, _ := charset.NewReader(response.Body, response.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}
	if response.StatusCode != 200 {
		return "", nil, fmt.Errorf(response.Status)
	}

	var result userQueryResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", nil, err
	}
	return result.AccountId, result.Location(), nil
}

func (api Api) User(user *pkg.User) (model.Account, *time.Location, error) {
	searchUrl := searchUserUrl
	searchPart := user.DisplayName
	if user.Id != "" {
		searchUrl = getUserUrl
		searchPart = user.Id
	}
	userUrl, err := api.Server.Parse(fmt.Sprintf(searchUrl, url.QueryEscape(searchPart)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, userUrl, api.Userinfo, nil)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	reader, _ := charset.NewReader(response.Body, response.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}
	if response.StatusCode != 200 {
		if response.StatusCode == 404 {
			return "", nil, fmt.Errorf("found no user for %s", user.DisplayName)
		}
		return "", nil, fmt.Errorf(response.Status)
	}

	if user.Id != "" {
		var result userQueryResult
		err = json.Unmarshal(data, &result)
		if err != nil {
			return "", nil, err
		}
		return result.AccountId, result.Location(), nil
	}

	var result = make([]userQueryResult, 0, 2)
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", nil, err
	}
	if len(result) == 0 || !user.Matches(pkg.User{DisplayName: result[0].DisplayName}) {
		return "", nil, fmt.Errorf("found no user for %s", user.DisplayName)
	}
	if len(result) > 1 && user.Matches(pkg.User{DisplayName: result[1].DisplayName}) {
		return "", nil, fmt.Errorf("found more than one user for %s", user.DisplayName)
	}
	return result[0].AccountId, result[0].Location(), nil
}

func (api Api) Issues(jql model.Jql, issueFunc model.IssueFunc) error {
	return api.issues(jql, 0, issueFunc)
}

func (api Api) issues(jql model.Jql, startAt int, issueFunc model.IssueFunc) error {
	body, _ := json.Marshal(issueQuery{
		Fields:         []string{"project"},
		Jql:            jql.Build(),
		PaginatedQuery: &PaginatedQuery{StartAt: startAt},
	})
	searchUrl, _ := api.Server.Parse(searchIssueUrl)
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodPost, searchUrl, api.Userinfo, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	reader, _ := charset.NewReader(response.Body, response.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf(response.Status)
	}

	var result = issueQueryResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	issues := result.issues()
	for _, e := range issues {
		issueFunc(e)
	}
	if (result.IsLast == nil && result.Total >= startAt+result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		err = api.issues(jql, startAt+result.MaxResults, issueFunc)
	}
	return err
}

func (api Api) Worklog(key model.IssueKey, worklogFunc model.WorklogFunc) error {
	return api.worklog(key, 0, worklogFunc)
}

func (api Api) worklog(key model.IssueKey, startAt int, worklogFunc model.WorklogFunc) error {
	worklogUrl, err := api.Server.Parse(fmt.Sprintf(getWorklogUrl, string(key), strconv.Itoa(startAt)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodGet, worklogUrl, api.Userinfo, nil)
	if err != nil {
		return err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	reader, _ := charset.NewReader(response.Body, response.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf(response.Status)
	}

	var result = worklogQueryResult{}
	err = json.Unmarshal(data, &result)
	items := result.worklogs()
	for _, e := range items {
		worklogFunc(e)
	}
	if (result.IsLast == nil && result.Total >= startAt+result.MaxResults) || (result.IsLast != nil && !*result.IsLast) {
		err = api.worklog(key, startAt+result.MaxResults, worklogFunc)
	}
	return err
}

func (api Api) AddWorklog(key model.IssueKey, date time.Time, duration time.Duration) error {
	body, _ := json.Marshal(worklogItem{
		Started:          date.Format(pkg.IsoDateTime),
		TimeSpentSeconds: int(duration.Truncate(time.Second).Seconds()),
	})
	worklogUrl, err := api.Server.Parse(fmt.Sprintf(addWorklogUrl, string(key)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodPost, worklogUrl, api.Userinfo, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	if response.StatusCode != 201 {
		return fmt.Errorf(response.Status)
	}

	return nil
}

func (api Api) RemoveWorklog(key model.IssueKey, id model.WorklogId) error {
	worklogUrl, err := api.Server.Parse(fmt.Sprintf(removeWorklogUrl, string(key), string(id)))
	response, err := pkg.CreateJsonRequest(api.Client, http.MethodDelete, worklogUrl, api.Userinfo, nil)
	if err != nil {
		return err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("Response could not be closed.", err)
		}
	}()

	if response.StatusCode != 204 {
		return fmt.Errorf(response.Status)
	}

	return nil
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

func (issue issue) Project() pkg.Project {
	return pkg.Project(issue.Fields.Project.Key)
}

func (issue issue) Key() model.IssueKey {
	return issue.ApiKey
}

func (issue issue) String() string {
	return fmt.Sprintf("%s;%s", issue.Project(), issue.Key())
}

func (author author) Id() model.Account {
	return author.AccountId
}

func (author author) String() string {
	return fmt.Sprintf("%s;%s;%s", author.AccountId, author.EmailAddress, author.DisplayName)
}

func (effort worklogItem) Id() model.WorklogId {
	return model.WorklogId(effort.ApiId)
}

func (effort worklogItem) Author() model.Author {
	return effort.ApiAuthor
}

func (effort worklogItem) Date() time.Time {
	date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
	// Do not convert the date to UTC first.
	// The user logs his effort in his current timezone, and this would lead to shifted time entries.
	//return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	//return date.Truncate(time.Hour * 24)
	return date
}

func (effort worklogItem) Comment() pkg.Description {
	description := ""
	if effort.ApiComment != nil && len(effort.ApiComment.Content) > 0 && len(effort.ApiComment.Content[0].Content) > 0 {
		description = effort.ApiComment.Content[0].Content[0].Text
	}
	return pkg.Description(description)
}

func (effort worklogItem) Duration() time.Duration {
	return time.Duration(effort.TimeSpentSeconds) * time.Second
}

func (effort worklogItem) String() string {
	return fmt.Sprintf("%s;%s;%s", effort.Date().Format(pkg.IsoYearMonthDay), effort.Duration(), effort.Comment())
}
