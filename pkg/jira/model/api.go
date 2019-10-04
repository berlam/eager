package model

import (
	"eager/pkg"
	"time"
)

type Account string

type IssueKey string

type Api interface {
	UserAccessor
	IssueAccessor
	WorklogAccessor
}

type UserAccessor interface {
	Me() (Account, *time.Location, error)
	User(user *pkg.User, projects []pkg.Project) (Account, *time.Location, error)
}

type IssueAccessor interface {
	Issues(jql Jql, startAt int) ([]Issue, error)
}

type WorklogAccessor interface {
	Worklog(key IssueKey, startAt int) ([]Worklog, error)
}

type Issue interface {
	Key() IssueKey
	Worklog(accounts map[Account]*pkg.User, worklog []Worklog, fromDate, toDate time.Time) map[Account]pkg.Timesheet
}

type Worklog interface {
	Author() Author
	Date() time.Time
	Comment() pkg.Description
	Duration() time.Duration
}

type Author interface {
	Id() Account
}
