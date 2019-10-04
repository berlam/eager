package jira

import (
	"eager/pkg"
	"fmt"
	"strings"
	"time"
)

const (
	jqlWorklogDate    = "worklogDate >= '%s' AND worklogDate < '%s'"
	jqlWorklogProject = "project in ('%s')"
	jqlWorklogAuthor  = "worklogAuthor in (%s)"
	currentUser       = "currentUser()"
)

type jql []string

func (query jql) me() jql {
	return jql(append(query, fmt.Sprintf(jqlWorklogAuthor, currentUser)))
}

func (query jql) projects(projects ...pkg.Project) jql {
	if projects == nil || len(projects) == 0 {
		return query
	}
	result := make([]string, len(projects))
	for i, project := range projects {
		result[i] = string(project)
	}
	return jql(append(query, fmt.Sprintf(jqlWorklogProject, strings.Join(result, "','"))))
}

func (query jql) users(users ...pkg.User) jql {
	if users == nil || len(users) == 0 {
		return query
	}
	result := make([]string, len(users))
	for i, user := range users {
		result[i] = string(user)
	}
	return jql(append(query, fmt.Sprintf(jqlWorklogAuthor, "'"+strings.Join(result, "','")+"'")))
}

func (query jql) between(fromDate, toDate time.Time) jql {
	return jql(append(query, fmt.Sprintf(jqlWorklogDate, fromDate.Format(pkg.IsoYearMonthDaySlash), toDate.Format(pkg.IsoYearMonthDaySlash))))
}

func (query jql) build() string {
	return strings.Join(query, " AND ")
}
