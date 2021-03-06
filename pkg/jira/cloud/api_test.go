// +build !unit

package cloud

import (
	"eager/pkg"
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
	user := &pkg.User{
		DisplayName: "Berla Atlassian Test User 2",
		TimeZone:    time.UTC,
	}
	account, _, e := api.User(user)
	if e != nil || account == "" {
		t.Error("User not found")
		return
	}
	//jql := model.Jql{}.Users(account)
	//e := api.Issues(jql, 0)
	//if e != nil || len(issues) == 0 {
	//	t.Error("Issues not found")
	//	return
	//}
	//e := api.Worklog(issues[0].Key(), 0)
	//if e != nil || len(worklog) == 0 {
	//	t.Error("Worklog not found")
	//	return
	//}
}
