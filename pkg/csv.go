package pkg

import (
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"strings"
	"time"
)

type CsvSpecification struct {
	fields      int
	header      bool
	user        *CsvProperty
	project     *CsvProperty
	task        *CsvProperty
	description *CsvProperty
	date        *CsvProperty
	duration    *CsvProperty
}

type CsvProperty struct {
	enabled bool
	index   int
}

func NewCsvSpecification() CsvSpecification {
	return CsvSpecification{
		fields:      0,
		header:      false,
		user:        newCsvProperty(),
		project:     newCsvProperty(),
		task:        newCsvProperty(),
		description: newCsvProperty(),
		date:        newCsvProperty(),
		duration:    newCsvProperty(),
	}
}

func newCsvProperty() *CsvProperty {
	return &CsvProperty{
		enabled: false,
		index:   0,
	}
}

func (spec CsvSpecification) Header(enable bool) CsvSpecification {
	spec.header = enable
	return spec
}

func (spec CsvSpecification) Skip() CsvSpecification {
	spec.fields++
	return spec
}

func (spec CsvSpecification) User(enable bool) CsvSpecification {
	if enable {
		spec.addField(spec.user)
	}
	return spec
}

func (spec CsvSpecification) Project(enable bool) CsvSpecification {
	if enable {
		spec.addField(spec.project)
	}
	return spec
}

func (spec CsvSpecification) Task(enable bool) CsvSpecification {
	if enable {
		spec.addField(spec.task)
	}
	return spec
}

func (spec CsvSpecification) Description(enable bool) CsvSpecification {
	if enable {
		spec.addField(spec.description)
	}
	return spec
}

func (spec CsvSpecification) Date(enable bool) CsvSpecification {
	if enable {
		spec.addField(spec.date)
	}
	return spec
}

func (spec CsvSpecification) Duration(enable bool) CsvSpecification {
	spec.addField(spec.duration)
	return spec
}

func (spec *CsvSpecification) addField(property *CsvProperty) {
	property.enabled = true
	property.index = spec.fields
	spec.fields++
}

func (ts Timesheet) ReadCsv(data []byte, spec *CsvSpecification) (Timesheet, error) {
	if len(data) == 0 {
		return ts, nil
	}

	csvr := csv.NewReader(bytes.NewReader(data))
	csvr.Comma = ';'
	csvr.ReuseRecord = true
	csvr.FieldsPerRecord = spec.fields

	if spec.header {
		_, err := csvr.Read()
		if err != nil {
			return ts, err
		}
	}
	// Guess that we have roundabout 20 items per month
	// for a single user
	timesheet := make(Timesheet, 0, 20)
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				return append(ts, timesheet...), nil
			}
			return ts, err
		}

		effort := Effort{}
		if spec.user.enabled {
			effort.User = User(row[spec.user.index])
		}
		if spec.project.enabled {
			effort.Project = Project(row[spec.project.index])
		}
		if spec.task.enabled {
			effort.Task = Task(row[spec.task.index])
		}
		if spec.description.enabled {
			effort.Description = Description(row[spec.description.index])
		}
		if spec.date.enabled {
			date, _ := time.Parse("02.01.2006", row[spec.date.index])
			effort.Date = date
		}
		if spec.duration.enabled {
			duration, _ := time.ParseDuration(strings.Replace(row[spec.duration.index], ",", ".", -1) + "h")
			effort.Duration = duration
		}
		timesheet = append(timesheet, effort)
	}
}

func (ts Timesheet) WriteCsv(writer io.Writer, spec *CsvSpecification, printEmptyLine bool) {
	if len(ts) == 0 {
		return
	}
	csvw := csv.NewWriter(writer)
	csvw.Comma = ';'

	var result = make([]string, spec.fields)

	if spec.header {
		if spec.user.enabled {
			result[spec.user.index] = "User"
		}
		if spec.project.enabled {
			result[spec.project.index] = "Project"
		}
		if spec.task.enabled {
			result[spec.task.index] = "Task"
		}
		if spec.description.enabled {
			result[spec.description.index] = "Description"
		}
		if spec.date.enabled {
			result[spec.date.index] = "Date"
		}
		if spec.duration.enabled {
			result[spec.duration.index] = "Duration"
		}
		err := csvw.Write(result)
		if err != nil {
			log.Println(err)
		}
	}

	date := time.Date(ts[0].Date.Year(), time.Month(ts[0].Date.Month()), 1, 0, 0, 0, 0, time.UTC)
	for _, effort := range ts {
		if printEmptyLine {
			for e := range result {
				result[e] = ""
			}
			for i := int(effort.Date.Sub(date).Truncate(time.Hour * 24).Hours() / 24); i > 0; i-- {
				if spec.user.enabled {
					result[spec.user.index] = string(effort.User)
				}
				if spec.date.enabled {
					result[spec.date.index] = date.Format(IsoYearMonthDay)
					date = date.Add(time.Hour * 24)
				}
				if spec.duration.enabled {
					var duration time.Duration
					duration = 0
					result[spec.duration.index] = duration.String()
				}
				err := csvw.Write(result)
				if err != nil {
					log.Println(err)
				}
			}
			date = date.Add(time.Hour * 24)
		}

		if spec.user.enabled {
			result[spec.user.index] = string(effort.User)
		}
		if spec.project.enabled {
			result[spec.project.index] = string(effort.Project)
		}
		if spec.task.enabled {
			result[spec.task.index] = string(effort.Task)
		}
		if spec.description.enabled {
			result[spec.description.index] = string(effort.Description)
		}
		if spec.date.enabled {
			result[spec.date.index] = effort.Date.Format(IsoYearMonthDay)
		}
		if spec.duration.enabled {
			result[spec.duration.index] = effort.Duration.String()
		}

		err := csvw.Write(result)
		if err != nil {
			log.Println(err)
		}
	}
	csvw.Flush()
}
