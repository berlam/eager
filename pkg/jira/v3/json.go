package v3

import "eager/pkg/jira/model"

type projectKey string

type PaginatedQuery struct {
	StartAt int `json:"startAt"`
}

type PaginatedResult struct {
	MaxResults int   `json:"maxResults"`
	StartAt    int   `json:"startAt"`
	Total      int   `json:"total"`
	IsLast     *bool `json:"isLast,omitempty"`
}

type projectQueryResult struct {
	*PaginatedResult
	Values []struct {
		ProjectKey projectKey `json:"key"`
	} `json:"values"`
}

type userQueryResult struct {
	AccountId model.Account `json:"accountId"`
}

type author struct {
	AccountId    model.Account `json:"accountId"`
	EmailAddress string        `json:"emailAddress"`
	DisplayName  string        `json:"displayName"`
}

type issue struct {
	Id     string         `json:"id"`
	ApiKey model.IssueKey `json:"key"`
	Fields struct {
		Project struct {
			Id   string     `json:"id"`
			Key  projectKey `json:"key"`
			Name string     `json:"name"`
		} `json:"project"`
	} `json:"fields"`
}

type issueQuery struct {
	*PaginatedQuery
	Fields []string `json:"fields"`
	Jql    string   `json:"jql"`
}

type issueQueryResult struct {
	*PaginatedResult
	ApiIssues []issue `json:"issues"`
}

type worklogQueryResult struct {
	*PaginatedResult
	ApiWorklogs []*worklogItem `json:"worklogs"`
}

type worklogItem struct {
	ApiAuthor    author `json:"author"`
	UpdateAuthor author `json:"updateAuthor"`
	ApiComment   struct {
		Type    string `json:"type"`
		Version int    `json:"version"`
		Content []*struct {
			Type    string `json:"type"`
			Content []*struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	} `json:"comment,omitempty"`
	Started          string `json:"started"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}
