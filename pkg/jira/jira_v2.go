package jira

import "eager/pkg"

type v2 struct {
}

func (v2) projects(startAt int) ([]pkg.Project, error) {
	panic("implement me")
}

func (v2) accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]Account, error) {
	panic("implement me")
}

func (v2) user(user pkg.User, projects []pkg.Project) (Account, error) {
	panic("implement me")
}

func (v2) issues(jql jql, startAt int) (Account, []Issue, error) {
	panic("implement me")
}

func (v2) worklog(key IssueKey, startAt int) ([]Worklog, error) {
	panic("implement me")
}
