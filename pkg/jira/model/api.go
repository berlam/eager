package model

import (
	"eager/pkg"
	"time"
)

const (
	HeaderAccountId = "X-AACCOUNTID"
)

type Account string

type IssueKey string

type Api interface {
	ProjectAccessor
	AccountAccessor
	UserAccessor
	IssueAccessor
	WorklogAccessor
}

type ProjectAccessor interface {
	Projects(startAt int) ([]pkg.Project, error)
}

type AccountAccessor interface {
	Accounts(projects []pkg.Project, users []pkg.User) (map[pkg.User]Account, error)
}

type UserAccessor interface {
	User(user pkg.User, projects []pkg.Project) (Account, error)
}

type IssueAccessor interface {
	Issues(jql Jql, startAt int) (Account, []Issue, error)
}

type WorklogAccessor interface {
	Worklog(key IssueKey, startAt int) ([]Worklog, error)
}

type Issue interface {
	Key() IssueKey
	Worklog(worklog []Worklog, fromDate, toDate time.Time) map[Account]pkg.Timesheet
}

type Worklog interface {
	IsBetween(fromDate, toDate time.Time) bool
	Author() Author
	Date() time.Time
	Comment() pkg.Description
	Duration() time.Duration
}

type Author interface {
	Id() Account
	Name() string
}
