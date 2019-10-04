package cloud

import (
	"eager/pkg"
	"eager/pkg/jira/model"
	v2 "eager/pkg/jira/v2"
	"net/http"
	"net/url"
)

const (
	BasePath         = "/rest/api/3/"
	searchProjectUrl = "project/search"
)

type Api struct {
	Client   *http.Client
	Server   *url.URL
	Userinfo *url.Userinfo
	v2       *v2.Api
}

func (api Api) previousVersion() *v2.Api {
	if api.v2 == nil {
		api.v2 = &v2.Api{
			Client:   api.Client,
			Server:   api.Server,
			Userinfo: api.Userinfo,
		}
	}
	return api.v2
}

func (api Api) Projects(startAt int) ([]pkg.Project, error) {
	return api.previousVersion().Projects(startAt)
}

func (api Api) User(user pkg.User, projects []pkg.Project) (model.Account, error) {
	return api.previousVersion().User(user, projects)
}

func (api Api) Issues(jql model.Jql, startAt int) (model.Account, []model.Issue, error) {
	return api.previousVersion().Issues(jql, startAt)
}

func (api Api) Worklog(key model.IssueKey, startAt int) ([]model.Worklog, error) {
	return api.previousVersion().Worklog(key, startAt)
}
