package internal

import (
	"eager/pkg"
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
	FlagSeconds        = "seconds"
	FlagAll            = "all"
	FlagForce          = "force"
)

func Projects(projects []string) []pkg.Project {
	result := make([]pkg.Project, len(projects))
	for i, project := range projects {
		result[i] = pkg.Project(project)
	}
	return result
}

func Users(users []string) []*pkg.User {
	result := make([]*pkg.User, len(users))
	for i, user := range users {
		result[i] = &pkg.User{
			DisplayName: user,
		}
	}
	return result
}

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
	Seconds        bool     `mapstructure:"seconds"`
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

func (c *Configuration) UsersArray() {

}
