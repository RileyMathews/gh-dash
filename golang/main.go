package main

import (
    "fmt"
    "os"
    "sync"
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

func fetchPR(url string, idx int, prs []PullRequest, wg *sync.WaitGroup) {
    defer wg.Done()
    var prDetail PrDetail
    if err := fetchGH(url, nil, &prDetail); err != nil {
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

    prs[idx] = fullPR
}

func fetchAllPRs(searchResult SearchIssuesResult) []PullRequest {
    n := len(searchResult.Items)
    prs := make([]PullRequest, n)
    var wg sync.WaitGroup
    wg.Add(n)

    for i, prItem := range searchResult.Items {
        go fetchPR(prItem.PullRequest.URL, i, prs, &wg)
    }

    wg.Wait()
    return prs
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

    var prs = fetchAllPRs(searchResult)

    for _, details := range prs {
        fmt.Printf("%+v\n", details)
    }
}

