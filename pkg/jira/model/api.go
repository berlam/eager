package model

import (
	"eager/pkg"
	"time"
)

type Account string

type IssueKey string

type WorklogId string

type Api interface {
	UserReader
	IssueReader
	WorklogReader
	WorklogWriter
}

type UserReader interface {
	Me() (Account, *time.Location, error)
	User(user *pkg.User) (Account, *time.Location, error)
}

type IssueReader interface {
	Issues(jql Jql, issueFunc IssueFunc) error
}

type IssueFunc func(Issue)

func (f IssueFunc) Process(issue Issue) {
	f(issue)
}

type WorklogReader interface {
	Worklog(key IssueKey, worklogFunc WorklogFunc) error
}

type WorklogWriter interface {
	AddWorklog(key IssueKey, date time.Time, duration time.Duration) error
	RemoveWorklog(key IssueKey, id WorklogId) error
}

type WorklogFunc func(Worklog) bool

type Issue interface {
	Project() pkg.Project
	Key() IssueKey
	String() string
}

type Worklog interface {
	Id() WorklogId
	Author() Author
	Date() time.Time
	Comment() pkg.Description
	Duration() time.Duration
	String() string
}

type Author interface {
	Id() Account
	String() string
}
