package pkg

import (
	"io"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Effort struct {
	User        *User
	Project     Project
	Task        Task
	Description Description
	Date        time.Time
	Duration    time.Duration
}

type User struct {
	DisplayName string
	TimeZone    *time.Location
}

type Project string

type Task string

type Description string

type Timesheet []Effort

func (user User) Matches(other User) bool {
	if user.TimeZone != nil && other.TimeZone != nil && user.TimeZone != other.TimeZone {
		return false
	}
	otherNormalized := other.Normalize()
	fields := strings.Fields(user.Normalize())
	for _, field := range fields {
		if !strings.Contains(otherNormalized, field) {
			return false
		}
	}
	return true
}

func (user User) Normalize() string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return -1
		}
		return r
	}, user.DisplayName)
}

func (ts Timesheet) sortByUserAndDateAndProjectAndTask() Timesheet {
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].User != nil && ts[j].User != nil && ts[i].User.DisplayName != ts[j].User.DisplayName {
			return ts[i].User.DisplayName < ts[j].User.DisplayName
		}
		if ts[i].Date != ts[j].Date {
			return ts[i].Date.Before(ts[j].Date)
		}
		if ts[i].Project != ts[j].Project {
			return ts[i].Project < ts[j].Project
		}
		if ts[i].Task != ts[j].Task {
			return ts[i].Task < ts[j].Task
		}
		return false
	})
	return ts
}

func (ts Timesheet) summarize() Timesheet {
	type Key struct {
		user string
		time.Time
	}
	sum := map[Key]*Effort{}
	for _, effort := range ts {
		name := ""
		if effort.User != nil {
			name = effort.User.DisplayName
		}
		key := Key{name, effort.Date}
		tmp := sum[key]
		if tmp == nil {
			sum[key] = &Effort{
				User:     effort.User,
				Date:     effort.Date,
				Duration: effort.Duration,
			}
			continue
		}
		tmp.Duration += effort.Duration
	}
	timesheet := make(Timesheet, len(sum))
	i := 0
	for _, effort := range sum {
		timesheet[i] = *effort
		i++
	}
	return timesheet
}

func (ts Timesheet) Print(writer io.Writer, user, summarize, printEmptyLine, seconds bool) {
	spec := NewCsvSpecification().User(user).Date(true).Project(!summarize).Task(!summarize).Duration(true).Description(!summarize)

	timesheet := ts
	if summarize {
		timesheet = timesheet.summarize()
	}
	timesheet.sortByUserAndDateAndProjectAndTask().WriteCsv(writer, &spec, printEmptyLine, seconds)
}
