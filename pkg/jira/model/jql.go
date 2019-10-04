package model

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
)

type Jql []string

func (query Jql) Projects(projects ...pkg.Project) Jql {
	if len(projects) == 0 {
		return query
	}
	result := make([]string, len(projects))
	for i, project := range projects {
		result[i] = string(project)
	}
	return Jql(append(query, fmt.Sprintf(jqlWorklogProject, strings.Join(result, "','"))))
}

func (query Jql) Users(users ...Account) Jql {
	if len(users) == 0 {
		return query
	}
	result := make([]string, len(users))
	for i, user := range users {
		result[i] = string(user)
	}
	return Jql(append(query, fmt.Sprintf(jqlWorklogAuthor, "'"+strings.Join(result, "','")+"'")))
}

func (query Jql) Between(fromDate, toDate time.Time) Jql {
	return Jql(append(query, fmt.Sprintf(jqlWorklogDate, fromDate.Format(pkg.IsoYearMonthDaySlash), toDate.Format(pkg.IsoYearMonthDaySlash))))
}

func (query Jql) Build() string {
	return strings.Join(query, " AND ")
}
