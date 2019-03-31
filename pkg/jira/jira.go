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
	currentUser        = "currentUser()"
	jqlWorklogDate     = "worklogDate >= '%s' AND worklogDate < '%s'"
	jqlWorklogProject  = "project in ('%s')"
	jqlWorklogAuthor   = "worklogAuthor in (%s)"
	jiraSearchIssueUrl = "/rest/api/3/search"
	jiraWorklogUrl     = "/rest/api/3/issue/%s/worklog?startAt=%s"
)

func GetTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects []string) pkg.Timesheet {
	fromDate, toDate := pkg.GetTimeRange(year, month)

	accountId, result, err := getJiraIssues(client, server, userinfo, fromDate, toDate, projects, nil)

	if err != nil {
		log.Println("Could not get Jira issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Effort)
	var wg sync.WaitGroup
	wg.Add(len(result.Issues))
	for _, item := range result.Issues {
		go func(wg *sync.WaitGroup, item issue) {
			defer wg.Done()
			worklogItems, err := getWorklogItems(client, server, userinfo, item.Key)
			if err != nil {
				log.Println("Could not get effort for "+item.Key, err)
			} else {
				getTimesheet(c, worklogItems, accountId, item.Fields.Project.Key, item.Key, fromDate, toDate)
			}
		}(&wg, item)
	}
	go func(wg *sync.WaitGroup) {
		defer close(c)
		wg.Wait()
	}(&wg)
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort)
	}

	return timesheet
}

func GetBulkTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, projects, users []string) pkg.Timesheet {
	fromDate, toDate := pkg.GetTimeRange(year, month)

	_, result, err := getJiraIssues(client, server, userinfo, fromDate, toDate, projects, users)

	if err != nil {
		log.Println("Could not get Jira issues.", err)
		return pkg.Timesheet{}
	}

	c := make(chan pkg.Effort)
	var wg sync.WaitGroup
	wg.Add(len(result.Issues))
	for _, item := range result.Issues {
		go func(wg *sync.WaitGroup, item issue) {
			defer wg.Done()
			worklogItems, err := getWorklogItems(client, server, userinfo, item.Key)
			if err != nil {
				log.Println("Could not get effort for "+item.Key, err)
			} else {
				items := worklogItems.getWorklogByEmailAddressAndDate(item.Fields.Project.Key, item.Key, fromDate, toDate)
				for _, user := range users {
					itemsOfUser := items[user]
					if itemsOfUser != nil {
						for _, itemOfUser := range itemsOfUser {
							c <- itemOfUser
						}
					}
				}
			}
		}(&wg, item)
	}
	go func(wg *sync.WaitGroup) {
		defer close(c)
		wg.Wait()
	}(&wg)
	var timesheet pkg.Timesheet
	for effort := range c {
		timesheet = append(timesheet, effort)
	}

	return timesheet
}

func (items worklogItems) getWorklogByEmailAddressAndDate(pKey projectKey, iKey issueKey, fromDate, toDate time.Time) map[string]pkg.Timesheet {
	result := make(map[string]pkg.Timesheet)
	total := len(items)
	for _, effort := range items {
		date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
		date = date.UTC()
		if !date.Before(fromDate) && date.Before(toDate) {
			current := result[effort.Author.EmailAddress]
			if current == nil {
				current = make(pkg.Timesheet, 0, total)
			}
			description := ""
			if len(effort.Comment.Content) > 0 && len(effort.Comment.Content[0].Content) > 0 {
				description = effort.Comment.Content[0].Content[0].Text
			}
			duration, _ := time.ParseDuration(strconv.Itoa(effort.TimeSpentSeconds) + "s")
			result[effort.Author.EmailAddress] = append(current, pkg.Effort{
				Employee:    pkg.Employee(effort.Author.DisplayName),
				Description: pkg.Description(description),
				Project:     pkg.Project(pKey),
				Task:        pkg.Task(iKey),
				Date:        time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
				Duration:    duration,
			})
		}
	}
	return result
}

func createRequest(client *http.Client, httpMethod string, server *url.URL, userinfo *url.Userinfo, payload io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(httpMethod, server.String(), payload)
	if err != nil {
		log.Fatal(err)
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

func getJiraIssues(client *http.Client, server *url.URL, userinfo *url.Userinfo, fromDate, toDate time.Time, projects, users []string) (accountId, *issueQueryResult, error) {
	jql := make([]string, 0, 3)
	jql = append(jql, fmt.Sprintf(jqlWorklogDate, fromDate.Format(pkg.IsoYearMonthDaySlash), toDate.Format(pkg.IsoYearMonthDaySlash)))
	if users != nil {
		jql = append(jql, fmt.Sprintf(jqlWorklogAuthor, "'"+strings.Join(users, "','")+"'"))
	} else {
		jql = append(jql, fmt.Sprintf(jqlWorklogAuthor, currentUser))
	}
	if projects != nil {
		jql = append(jql, fmt.Sprintf(jqlWorklogProject, strings.Join(projects, "','")))
	}
	body, _ := json.Marshal(issueQuery{
		Fields:         []string{"project"},
		Jql:            strings.Join(jql, " AND "),
		PaginatedQuery: &PaginatedQuery{StartAt: 0},
	})
	searchUrl, _ := server.Parse(jiraSearchIssueUrl)
	response, err := createRequest(client, http.MethodPost, searchUrl, userinfo, bytes.NewBuffer(body))
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	account, _ := url.PathUnescape(response.Header.Get("X-AACCOUNTID"))
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil, err
	}
	if response.StatusCode != 200 {
		return "", nil, fmt.Errorf(response.Status)
	}

	var result = issueQueryResult{}
	err = json.Unmarshal(data, &result)
	return accountId(account), &result, err
}

func getWorklogItems(client *http.Client, server *url.URL, userinfo *url.Userinfo, key issueKey) (worklogItems, error) {
	worklogUrl, err := server.Parse(fmt.Sprintf(jiraWorklogUrl, string(key), strconv.Itoa(0)))
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
	return result.Worklogs, nil
}

func getTimesheet(c chan pkg.Effort, worklogItems []*worklogItem, accountId accountId, pKey projectKey, iKey issueKey, fromDate, toDate time.Time) {
	var wg sync.WaitGroup
	wg.Add(len(worklogItems))
	for _, effort := range worklogItems {
		go func(wg *sync.WaitGroup, effort *worklogItem) {
			defer wg.Done()
			date, _ := time.Parse(pkg.IsoDateTime, effort.Started)
			date = date.UTC()
			if effort.Author.AccountId == accountId && !date.Before(fromDate) && date.Before(toDate) {
				description := ""
				if len(effort.Comment.Content) > 0 && len(effort.Comment.Content[0].Content) > 0 {
					description = effort.Comment.Content[0].Content[0].Text
				}
				duration, _ := time.ParseDuration(strconv.Itoa(effort.TimeSpentSeconds) + "s")
				c <- pkg.Effort{
					Description: pkg.Description(description),
					Project:     pkg.Project(pKey),
					Task:        pkg.Task(iKey),
					Date:        time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
					Duration:    duration,
				}
			}
		}(&wg, effort)
	}
	wg.Wait()
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
	MaxResults int  `json:"maxResults"`
	StartAt    int  `json:"startAt"`
	Total      int  `json:"total"`
	IsLast     bool `json:"isLast"`
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
