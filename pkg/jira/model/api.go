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
	User(user *pkg.User) (Account, *time.Location, error)
}

type IssueAccessor interface {
	Issues(jql Jql, issueFunc IssueFunc) error
}

type IssueFunc func(Issue)

func (f IssueFunc) Process(issue Issue) {
	f(issue)
}

type WorklogAccessor interface {
	Worklog(key IssueKey, worklogFunc WorklogFunc) error
}

type WorklogFunc func(Worklog)

func (f WorklogFunc) Process(worklog Worklog) {
	f(worklog)
}

type Issue interface {
	Project() pkg.Project
	Key() IssueKey
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
