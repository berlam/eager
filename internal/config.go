package internal

import (
	"net/url"
)

const (
	Version            = "0.1.0"
	FlagInsecure       = "insecure"
	FlagHost           = "host"
	FlagUsername       = "username"
	FlagPassword       = "password"
	FlagYear           = "year"
	FlagMonth          = "month"
	FlagProjects       = "project"
	FlagUsers          = "user"
	FlagReport         = "report"
	FlagSummarize      = "summarize"
	FlagPrintEmptyLine = "printemptyline"
	FlagAll            = "all"
	FlagForce          = "force"
)

type Configuration struct {
	Insecure       bool     `mapstructure:"insecure"`
	Host           string   `mapstructure:"host"`
	Username       string   `mapstructure:"username"`
	Password       string   `mapstructure:"password"`
	Year           int      `mapstructure:"year"`
	Month          int      `mapstructure:"month"`
	Projects       []string `mapstructure:"projects"`
	Users          []string `mapstructure:"users"`
	Report         string   `mapstructure:"report"`
	Summarize      bool     `mapstructure:"summarize"`
	PrintEmptyLine bool     `mapstructure:"printemptyline"`
}

var Config Configuration

func (c *Configuration) Server() *url.URL {
	scheme := "https"
	if c.Insecure {
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
