package pkg

import (
	"io"
	"sort"
	"time"
)

type Effort struct {
	Employee    Employee
	Project     Project
	Task        Task
	Description Description
	Date        time.Time
	Duration    time.Duration
}

type Employee string

type Project string

type Task string

type Description string

type Timesheet []Effort

func (ts Timesheet) sortByEmployeeAndDateAndProjectAndTask() Timesheet {
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].Employee != ts[j].Employee {
			return ts[i].Employee < ts[j].Employee
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
		Employee
		time.Time
	}
	sum := map[Key]*Effort{}
	for _, effort := range ts {
		tmp := sum[Key{effort.Employee, effort.Date}]
		if tmp == nil {
			sum[Key{effort.Employee, effort.Date}] = &Effort{
				Employee: effort.Employee,
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

func (ts Timesheet) Print(writer io.Writer, employee, summarize, printEmptyLine bool) {
	spec := NewCsvSpecification().Employee(employee).Date(true).Project(!summarize).Task(!summarize).Duration(true).Description(!summarize)

	timesheet := ts
	if summarize {
		timesheet = timesheet.summarize()
	}
	timesheet.sortByEmployeeAndDateAndProjectAndTask().WriteCsv(writer, &spec, printEmptyLine)
}
