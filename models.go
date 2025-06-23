package main

import "time"

// Define data models corresponding to GitHub API response structures.

// Basic user information
type GithubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Type  string `json:"type"`
}

// Repository information (owner is a GithubUser)
type GithubRepo struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	FullName string     `json:"full_name"`
	Private  bool       `json:"private"`
	Owner    GithubUser `json:"owner"`
}

type GithubUrl struct {
	Href string `json:"href"`
}

type PrLinks struct {
	Self GithubUrl `json:"self"`
	Html GithubUrl `json:"html"`
}

type PrReview struct {
	User        GithubUser `json:"user"`
	State       string     `json:"state"`
	SubmittedAt time.Time  `json:"submitted_at"`
}

// Pull Request detail (basic PR info and links to related resources)
type PrDetail struct {
	Title  string `json:"title"`
	State  string `json:"state"`
	Number int    `json:"number"`
	Base   struct {
		Repo GithubRepo `json:"repo"`
	} `json:"base"`
	CommitsURL  string     `json:"commits_url"`
	CommentsURL string     `json:"comments_url"`
	CreatedAt   time.Time  `json:"created_at"`
	User        GithubUser `json:"user"`
	UiUrl       string     `json:"issue_url"`
	Links       PrLinks    `json:"_links"`
	Draft       bool       `json:"draft"`
	Head        struct {
		Ref  string     `json:"ref"`
		Repo GithubRepo `json:"repo"`
	} `json:"head"`
}

type PrCheckRunsResponse struct {
	CheckRuns []struct {
		Conclusion string `json:"conclusion"`
		Status     string `json:"status"`
	} `json:"check_runs"`
}

// Comment on a PR (we capture the author and creation time)
type Comment struct {
	User      GithubUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
}

// Commit holds commit metadata (we include only the author sub-structure here)
type Commit struct {
	Author struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Date  time.Time `json:"date"`
	} `json:"author"`
}

// PR commit entry, including the GitHub user (author) and the commit details
type PrCommit struct {
	Author GithubUser `json:"author"`
	Commit Commit     `json:"commit"`
}

// Aggregate PullRequest structure combining PR details, comments, and commits
type PullRequest struct {
	Pr       PrDetail            `json:"pr"`
	Comments []Comment           `json:"comments"`
	Commits  []PrCommit          `json:"commits"`
	Reviews  []PrReview          `json:"reviews"`
	Checks   PrCheckRunsResponse `json:"check_runs_response"`
}

type Section struct {
	Name         string
	PullRequests []ProcessedPr
}

type Action struct {
	User string
	Type string
	Time time.Time
}

type ProcessedPr struct {
	RawPr              PullRequest
	Actions            []Action
	Reviewed           bool
	ChangeRequested    bool
	Approved           bool
	ChecksStillRunning bool
	ChecksFailed       bool
}
