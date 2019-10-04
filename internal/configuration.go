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
	FlagYear          = "year"
	FlagMonth         = "month"
	FlagProjects      = "project"
	FlagUsers         = "user"
	FlagReport        = "report"
	FlagSummarize     = "summarize"
	FlagEmpty         = "empty"
	FlagDecimal       = "decimal"
	FlagNegate        = "negate"
	FlagAll           = "all"
	FlagForce         = "force"
)

type Configuration struct {
	Http     bool            `mapstructure:"http"`
	Host     string          `mapstructure:"host"`
	Username string          `mapstructure:"username"`
	Password string          `mapstructure:"password"`
	Year     int             `mapstructure:"year"`
	Month    int             `mapstructure:"month"`
	Projects []string        `mapstructure:"projects"`
	Users    []string        `mapstructure:"users"`
	Report   string          `mapstructure:"report"`
	Duration DurationOptions `mapstructure:",squash"`
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
