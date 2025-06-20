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

func fetchPR(url string) PullRequest {
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

    return fullPR
}

func fetchAllPRs(searchResult SearchIssuesResult) []PullRequest {
    var (
        prs []PullRequest
        mu  sync.Mutex       // protects prs slice
        wg  sync.WaitGroup   // waits for all fetches
    )

    // Pre-allocate if you know the count
    prs = make([]PullRequest, 0, len(searchResult.Items))

    wg.Add(len(searchResult.Items))             // 1
    for _, pr := range searchResult.Items {
        go func(prItem SearchItem) {                // 2
            defer wg.Done()                     // 3

            details := fetchPR(prItem.PullRequest.URL)
            mu.Lock()                           // 4
            prs = append(prs, details)
            mu.Unlock()
        }(pr)
    }

    wg.Wait()                                   // 5
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

