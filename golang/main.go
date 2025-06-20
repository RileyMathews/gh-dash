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

func getLastActionUser(pr PullRequest) string {
	var newestComment *Comment
	for _, comment := range pr.Comments {
		if newestComment == nil || comment.CreatedAt.After(newestComment.CreatedAt) {
			newestComment = &comment
		}
	}

	var newestCommit *PrCommit
	for _, commit := range pr.Commits {
		if newestCommit == nil || commit.Commit.Author.Date.After(newestCommit.Commit.Author.Date) {
			newestCommit = &commit
		}
	}

	if newestComment != nil {
		return newestComment.User.Login
	}

	if newestCommit != nil {
		return newestCommit.Author.Login
	}

	return pr.Pr.User.Login
}

func sortPRs(prs []PullRequest) []Section {
	var myPrsNeedAttention = Section{
		Name:         "My Prs that need attention",
		PullRequests: []PullRequest{},
	}
	var myPrsNoActionNeeded = Section{
		Name:         "My Prs no action needed",
		PullRequests: []PullRequest{},
	}

	for _, pr := range prs {
		var lastActor = getLastActionUser(pr)
		if pr.Pr.User.Login == "RileyMathews" {
			if lastActor != "RileyMathews" {
				myPrsNeedAttention.PullRequests = append(myPrsNeedAttention.PullRequests, pr)
			} else {
				myPrsNoActionNeeded.PullRequests = append(myPrsNoActionNeeded.PullRequests, pr)
			}
		}
	}

	return []Section{myPrsNeedAttention, myPrsNoActionNeeded}
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
	var sections = sortPRs(prs)

	for _, section := range sections {
		fmt.Printf("%s\n", section.Name)
		for _, pr := range section.PullRequests {
			fmt.Printf("%s %s\n", pr.Pr.Title, pr.Pr.UiUrl)
		}
		fmt.Printf("\n")
	}
}
