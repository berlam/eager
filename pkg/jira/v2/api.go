package v2

import (
	"eager/pkg"
	"eager/pkg/jira/model"
	"net/http"
	"net/url"
)

const (
	searchProjectUrl = "/rest/api/2/project/search?startAt=%s"
	searchUserUrl    = "/rest/api/2/user/assignable/multiProjectSearch?projectKeys=%s&query=%s&maxResults=2"
	searchIssueUrl   = "/rest/api/2/search"
	worklogUrl       = "/rest/api/2/issue/%s/worklog?startAt=%s"
)

type Api struct {
	Client   *http.Client
	Server   *url.URL
	Userinfo *url.Userinfo
}

func (Api) Projects(startAt int) ([]pkg.Project, error) {
	panic("implement me")
}

func (Api) Accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]model.Account, error) {
	panic("implement me")
}

func (Api) User(user pkg.User, projects []pkg.Project) (model.Account, error) {
	panic("implement me")
}

func (Api) Issues(jql model.Jql, startAt int) (model.Account, []model.Issue, error) {
	panic("implement me")
}

func (Api) Worklog(key model.IssueKey, startAt int) ([]model.Worklog, error) {
	panic("implement me")
}
