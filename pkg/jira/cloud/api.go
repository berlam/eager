package cloud

import (
	"eager/pkg"
	"eager/pkg/jira/model"
	"eager/pkg/jira/v2"
	"net/http"
	"net/url"
	"time"
)

const (
	BasePath = "/rest/api/3/"
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

func (api Api) Me() (model.Account, *time.Location, error) {
	return api.previousVersion().Me()
}

func (api Api) User(user *pkg.User) (model.Account, *time.Location, error) {
	return api.previousVersion().User(user)
}

func (api Api) Issues(jql model.Jql, issueFunc model.IssueFunc) error {
	return api.previousVersion().Issues(jql, issueFunc)
}

func (api Api) Worklog(key model.IssueKey, worklogFunc model.WorklogFunc) error {
	return api.previousVersion().Worklog(key, worklogFunc)
}

func (api Api) AddWorklog(key model.IssueKey, date time.Time, duration time.Duration) error {
	return api.previousVersion().AddWorklog(key, date, duration)
}

func (api Api) RemoveWorklog(key model.IssueKey, id model.WorklogId) error {
	return api.previousVersion().RemoveWorklog(key, id)
}
