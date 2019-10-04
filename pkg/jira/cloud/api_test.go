// +build !unit

package cloud

import (
	"eager/pkg"
	"eager/pkg/jira/model"
	"net/url"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	cloudServer, _ := os.LookupEnv("EAGER_ATLASSIAN_CLOUD_URL")
	cloudUser, _ := os.LookupEnv("EAGER_ATLASSIAN_CLOUD_USER")
	cloudToken, _ := os.LookupEnv("EAGER_ATLASSIAN_CLOUD_TOKEN")
	path, _ := url.Parse(cloudServer)
	path, _ = path.Parse(BasePath)
	api := Api{
		Client:   pkg.NewHttpClient(),
		Server:   path,
		Userinfo: url.UserPassword(cloudUser, cloudToken),
	}
	projects, e := api.Projects(0)
	if e != nil || len(projects) == 0 || projects[0] != "TEST" {
		t.Error("Project not found")
		return
	}
	user := &pkg.User{
		DisplayName: "Berla Atlassian Test User 2",
		TimeZone:    time.UTC,
	}
	account, _, e := api.User(user, projects)
	if e != nil || account == "" {
		t.Error("User not found")
		return
	}
	jql := model.Jql{}.Users(account).Projects(projects...)
	issues, e := api.Issues(jql, 0)
	if e != nil || len(issues) == 0 {
		t.Error("Issues not found")
		return
	}
	worklog, e := api.Worklog(issues[0].Key(), 0)
	if e != nil || len(worklog) == 0 {
		t.Error("Worklog not found")
		return
	}
}
