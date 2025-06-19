package main

import (
    "encoding/json"
    "fmt"
    "os"
)

// Data structures for parsing search results
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
    // Define search parameters (equivalent to the Python search_params dict)
    searchParams := map[string]string{
        "q":        "is:pr is:open org:MercuryTechnologies author:@me",
        "per_page": "10",
    }

    // Fetch search results for open PRs
    data := fetchGH("/search/issues", searchParams)
    var searchResult SearchIssuesResult
    if err := json.Unmarshal(data, &searchResult); err != nil {
        fmt.Printf("Error decoding search results: %v\n", err)
        os.Exit(1)
    }
    if len(searchResult.Items) == 0 {
        // No pull requests found; exit with code 0
        os.Exit(0)
    }

    // Take the first PR from results and get its API URL
    prURL := searchResult.Items[0].PullRequest.URL
    if prURL == "" {
        // If for some reason the pull_request URL is missing, exit with error
        os.Exit(1)
    }

    // Fetch detailed PR data
    data = fetchGH(prURL, nil)
    var prDetail PrDetail
    if err := json.Unmarshal(data, &prDetail); err != nil {
        fmt.Printf("Error decoding PR details: %v\n", err)
        os.Exit(1)
    }

    // Fetch comments for the PR
    data = fetchGH(prDetail.CommentsURL, nil)
    var comments []Comment
    if err := json.Unmarshal(data, &comments); err != nil {
        fmt.Printf("Error decoding comments: %v\n", err)
        os.Exit(1)
    }

    // Fetch commits for the PR
    data = fetchGH(prDetail.CommitsURL, nil)
    var commits []PrCommit
    if err := json.Unmarshal(data, &commits); err != nil {
        fmt.Printf("Error decoding commits: %v\n", err)
        os.Exit(1)
    }

    // Combine PR details, comments, and commits into a single structure
    fullPR := PullRequest{
        Pr:       prDetail,
        Comments: comments,
        Commits:  commits,
    }

    // Print the full PullRequest data structure
    fmt.Printf("%+v\n", fullPR)
}

