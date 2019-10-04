// +build !unit

package v2

import (
	"bytes"
	"context"
	"eager/pkg"
	"eager/pkg/jira/model"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/headzoo/surf/browser"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/headzoo/surf.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

const port = "8080"

func Test(t *testing.T) {
	ctx := context.Background()
	jira, err := setupContainer(&ctx)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = jira.Terminate(ctx)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	host, err := jira.Host(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	port, err := jira.MappedPort(ctx, port)
	if err != nil {
		t.Error(err)
		return
	}

	err = setupJira(host, port)
	if err != nil {
		t.Error(err)
		return
	}

	tests := []struct {
		name string
	}{
		{name: "Add worklog"},
		{name: "Get worklog"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

func addProject(host string, port nat.Port) projectKey {
	client := pkg.NewHttpClient()
	server, _ := url.Parse(host + ":" + port.Port() + "/rest/api/2/project")
	type request struct {
		Key            projectKey `json:"key"`
		Name           string     `json:"name"`
		Lead           string     `json:"lead"`
		ProjectTypeKey string     `json:"projectTypeKey"`
	}
	body, _ := json.Marshal(request{
		Key:            "TEST",
		Name:           "Test",
		Lead:           "admin",
		ProjectTypeKey: "business",
	})
	response, _ := pkg.CreateJsonRequest(client, http.MethodPost, server, url.UserPassword("admin", "admin"), bytes.NewBuffer(body))
	defer func() {
		_ = response.Body.Close()
	}()
	var result struct {
		Key projectKey `json:"key"`
	}
	data, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(data, &result)
	return result.Key
}

func addIssue(host string, port nat.Port, key projectKey) model.IssueKey {
	client := pkg.NewHttpClient()
	server, _ := url.Parse(host + ":" + port.Port() + "/rest/api/2/issue")
	type id struct {
		Id string `json:"id"`
	}
	type fields struct {
		Project   id     `json:"project"`
		Summary   string `json:"summary"`
		IssueType id     `json:"issuetype"`
	}
	type request struct {
		Fields fields `json:"fields"`
	}
	payload := &request{
		Fields: fields{
			Project: id{
				Id: string(key),
			},
			Summary: "Issue 1",
			IssueType: id{
				Id: "10000",
			},
		},
	}

	body, _ := json.Marshal(payload)
	response, _ := pkg.CreateJsonRequest(client, http.MethodPost, server, url.UserPassword("admin", "admin"), bytes.NewBuffer(body))
	defer func() {
		_ = response.Body.Close()
	}()
	var result struct {
		Key model.IssueKey `json:"key"`
	}
	data, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(data, &result)
	return result.Key
}

func setupContainer(ctx *context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "dchevell/jira-core@sha256:72aba213c4974002a5969a5ff0e72480a51cb8ff31fa15530dff854cf2350313",
		ExposedPorts: []string{port + "/tcp"},
		WaitingFor:   wait.ForLog("LauncherContextListener"),
	}
	container, err := testcontainers.GenericContainer(*ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func setupJira(host string, port nat.Port) error {
	// Create a new browser and open jira.
	bow := surf.NewBrowser()
	err := bow.Open(fmt.Sprintf("http://%s:%s", host, port.Port()))
	if err != nil {
		return err
	}

	// Advanced settings mode
	err = submit(bow, "#jira-setup-mode", map[string]string{
		"setupOption": "classic",
	})
	if err != nil {
		return err
	}

	// Database
	err = submit(bow, "#jira-setup-database", map[string]string{
		"databaseOption": "internal",
	})
	if err != nil {
		return err
	}

	// General settings
	err = submit(bow, "#jira-setupwizard", map[string]string{
		"title":   "Jira",
		"mode":    "private",
		"baseURL": "http://localhost:8080",
	})
	if err != nil {
		return err
	}

	// License key
	// https://developer.atlassian.com/platform/marketplace/timebomb-licenses-for-testing-server-apps/
	license, _ := ioutil.ReadFile("test-fixtures/license.txt")
	err = submit(bow, "#setupLicenseForm", map[string]string{
		"setupLicenseKey": string(license),
	})
	if err != nil {
		return err
	}

	// Admin user
	err = submit(bow, "#jira-setupwizard", map[string]string{
		"fullname": "Administrator",
		"email":    "john.doe@example.org",
		"username": "admin",
		"password": "admin",
		"confirm":  "admin",
	})
	if err != nil {
		return err
	}

	// Email notification
	err = submit(bow, "#jira-setupwizard", map[string]string{
		"noemail": "true",
	})
	if err != nil {
		return err
	}

	return nil
}

func submit(bow *browser.Browser, selector string, values map[string]string) error {
	form, err := bow.Form(selector)
	if err != nil {
		return err
	}
	for k, v := range values {
		err = form.Input(k, v)
		if err != nil {
			return err
		}
	}
	err = form.Submit()
	if err != nil {
		return err
	}
	return nil
}
