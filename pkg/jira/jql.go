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

func (query jql) projects(projects ...string) jql {
	if projects == nil {
		return query
	}
	return jql(append(query, fmt.Sprintf(jqlWorklogProject, strings.Join(projects, "','"))))
}

func (query jql) users(users ...string) jql {
	if users == nil {
		return query
	}
	return jql(append(query, fmt.Sprintf(jqlWorklogAuthor, "'"+strings.Join(users, "','")+"'")))
}

func (query jql) between(fromDate, toDate time.Time) jql {
	return jql(append(query, fmt.Sprintf(jqlWorklogDate, fromDate.Format(pkg.IsoYearMonthDaySlash), toDate.Format(pkg.IsoYearMonthDaySlash))))
}

func (query jql) build() string {
	return strings.Join(query, " AND ")
}
