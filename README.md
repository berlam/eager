eager
=====
```Shell
Eager is a tool for real eager beavers to maintain and synchronize one worklog across different services.

Usage:
   [command]

Available Commands:
  help        Help about any command
  show        Show worklog

Flags:
      --config string     specify the configuration file
  -h, --help              help for this command
  -H, --host string       specify the host to use for effort query
  -k, --insecure          use http instead of https
  -p, --password string   specify the password to use for server authentication
  -u, --username string   specify the username to use for server authentication
      --version           version for this command

Use " [command] --help" for more information about a command.
```
### Output ###
```Shell
$DATE;$PROJECT;$TASK;$DURATION;$DESCRIPTION
[...]
```

### [Atlassian Jira](https://www.atlassian.com/software/jira/) ###
- [API v2](https://developer.atlassian.com/cloud/jira/platform/rest/v2/)
- [API v3](https://developer.atlassian.com/cloud/jira/platform/rest/v3/)

API version 3 with basic auth supported.

Time tracking must be enabled in Jira.

Permissions required:
- Browse projects project permission for the project that the issue is in.
- If configured, permission to see the issue granted by issue-level security.

Only worklogs visible to the user are returned. Visible worklogs are those:
- that have no visibility restrictions set.
- where the user belongs to the group or has the role visibility is restricted to.

### TODO [Bugzilla](https://www.bugzilla.org/) ###
- [API](https://www.bugzilla.org/docs/4.4/en/html/api/Bugzilla/WebService/Bug.html)
- [Docker](https://hub.docker.com/u/bugzilla/)

### TODO [GitLab](https://www.gitlab.com/) ###
- [API](https://docs.gitlab.com/ee/workflow/time_tracking.html)
- [Docker](https://docs.gitlab.com/omnibus/docker/)

### TODO [JetBrains YouTrack](https://www.jetbrains.com/youtrack/) ###
- [API](https://www.jetbrains.com/help/youtrack/standalone/Time-Tracking-REST-API.html)
- [Docker](https://hub.docker.com/r/jetbrains/youtrack/)

### [Projektron BCS](https://www.projektron.de/bcs/) ###
Create a effort filter inside the web application with following columns enabled:
- Project
- Task
- Date
- Duration

Set the display format for effort to decimal.

You might also filter for your project or every other property you like inside that filter.
Every filter result listed there will be used for the worklog.

### TODO [Redmine](https://www.redmine.org/) ###
- [API](https://www.redmine.org/projects/redmine/wiki/RedmineTimeTracking)
- [Docker](https://hub.docker.com/_/redmine)

## Build from source ##
If you want to build it right away you need to have a working [Go environment](https://golang.org/doc/install).
```Shell
$ go get -u
$ go install
```

## Credits ##

This project was created by [@berlam](https://github.com/berlam).

## License ##

See [LICENSE](LICENSE).
