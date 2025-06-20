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
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    FullName string    `json:"full_name"`
    Private  bool      `json:"private"`
    Owner    GithubUser `json:"owner"`
}

// Base contains the repository info for a PR (part of PR details)
type Base struct {
    Repo GithubRepo `json:"repo"`
}

// Pull Request detail (basic PR info and links to related resources)
type PrDetail struct {
    Title       string    `json:"title"`
    Number      int       `json:"number"`
    Base        Base      `json:"base"`
    CommitsURL  string    `json:"commits_url"`
    CommentsURL string    `json:"comments_url"`
    CreatedAt   time.Time `json:"created_at"`
}

// Comment on a PR (we capture the author and creation time)
type Comment struct {
    User      GithubUser `json:"user"`
    CreatedAt time.Time  `json:"created_at"`
}

// Commit author details (from the commit object in a PR commit)
type CommitAuthor struct {
    Name  string    `json:"name"`
    Email string    `json:"email"`
    Date  time.Time `json:"date"`
}

// Commit holds commit metadata (we include only the author sub-structure here)
type Commit struct {
    Author CommitAuthor `json:"author"`
}

// PR commit entry, including the GitHub user (author) and the commit details
type PrCommit struct {
    Author GithubUser `json:"author"`
    Commit Commit     `json:"commit"`
}

// Aggregate PullRequest structure combining PR details, comments, and commits
type PullRequest struct {
    Pr       PrDetail   `json:"pr"`
    Comments []Comment  `json:"comments"`
    Commits  []PrCommit `json:"commits"`
}

