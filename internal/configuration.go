package internal

import (
	"net/url"
)

const (
	Version           = "0.1.0"
	FlagConfiguration = "configuration"
	FlagHttp          = "http"
	FlagHost          = "host"
	FlagUsername      = "username"
	FlagPassword      = "password"
	FlagProjects      = "project"
	FlagUsers         = "user"
	FlagReport        = "report"
	FlagSummarize     = "summarize"
	FlagEmpty         = "empty"
	FlagDecimal       = "decimal"
	FlagNegate        = "negate"
	FlagYear          = "year"
	FlagMonth         = "month"
	FlagDay           = "day"
	FlagTask          = "task"
)

type Configuration struct {
	Http     bool            `mapstructure:"http"`
	Host     string          `mapstructure:"host"`
	Username string          `mapstructure:"username"`
	Password string          `mapstructure:"password"`
	Projects []string        `mapstructure:"projects"`
	Users    []string        `mapstructure:"users"`
	Report   string          `mapstructure:"report"`
	Duration DurationOptions `mapstructure:",squash"`
	// These items make no sense to have inside a configuration file
	Year  int
	Month int
	Day   int
	Task  string
}

type DurationOptions struct {
	Summarize bool `mapstructure:"summarize"`
	Empty     bool `mapstructure:"empty"`
	Decimal   bool `mapstructure:"decimal"`
	Negate    bool `mapstructure:"negate"`
}

func (c *Configuration) Server() *url.URL {
	scheme := "https"
	if c.Http {
		scheme = "http"
	}
	return &url.URL{
		Scheme: scheme,
		Host:   c.Host,
	}
}

func (c *Configuration) Userinfo() *url.Userinfo {
	if c.Username != "" {
		if c.Password != "" {
			return url.UserPassword(c.Username, c.Password)
		}
		return url.User(c.Username)
	}
	return nil
}
