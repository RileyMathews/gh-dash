package main

import (
    "fmt"
    "os"
)

// Search result types (unchanged)
type SearchPullRequestInfo struct {
    URL string `json:"url"`
}
type SearchItem struct {
    PullRequest SearchPullRequestInfo `json:"pull_request"`
}
type SearchIssuesResult struct {
    Items []SearchItem `json:"items"`
}

func main() {
    searchParams := map[string]string{
        "q":        "is:pr is:open org:MercuryTechnologies author:@me",
        "per_page": "10",
    }

    // 1) Fetch open PR listings directly into searchResult
    var searchResult SearchIssuesResult
    if err := fetchGH("/search/issues", searchParams, &searchResult); err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching PR list: %v\n", err)
        os.Exit(1)
    }
    if len(searchResult.Items) == 0 {
        os.Exit(0)
    }

    prURL := searchResult.Items[0].PullRequest.URL
    if prURL == "" {
        os.Exit(1)
    }

    // 2) Fetch PR details
    var prDetail PrDetail
    if err := fetchGH(prURL, nil, &prDetail); err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching PR details: %v\n", err)
        os.Exit(1)
    }

    // 3) Fetch comments
    var comments []Comment
    if err := fetchGH(prDetail.CommentsURL, nil, &comments); err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching comments: %v\n", err)
        os.Exit(1)
    }

    // 4) Fetch commits
    var commits []PrCommit
    if err := fetchGH(prDetail.CommitsURL, nil, &commits); err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching commits: %v\n", err)
        os.Exit(1)
    }

    fullPR := PullRequest{
        Pr:       prDetail,
        Comments: comments,
        Commits:  commits,
    }

    // Print combined PR info
    fmt.Printf("%+v\n", fullPR)
}

