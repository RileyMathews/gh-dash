package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	RED       = "\033[31m"
	GREEN     = "\033[32m"
	YELLOW    = "\033[33m"
	BLUE      = "\033[34m"
	RESET     = "\033[0m" // Reset to default color
	CHECKMARK = ""
	X         = ""
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

	var reviews []PrReview
	reviewsUrl := fmt.Sprintf("%s/reviews", prDetail.Links.Self.Href)
	if err := fetchGH(reviewsUrl, nil, &reviews); err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching commits: %v\n", err)
		os.Exit(1)
	}

	var checks PrCheckRunsResponse
	checksUrl := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/check-runs", prDetail.Base.Repo.FullName, prDetail.Head.Ref)
	if err := fetchGH(checksUrl, nil, &checks); err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching commits: %v\n", err)
		os.Exit(1)
	}

	fullPR := PullRequest{
		Pr:       prDetail,
		Comments: comments,
		Commits:  commits,
		Reviews:  reviews,
		Checks:   checks,
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

func getActions(pr PullRequest) []Action {
	var newestComment *Comment
	for _, comment := range pr.Comments {
		if strings.Contains(comment.User.Login, "[bot]") {
			continue
		}
		if newestComment == nil || comment.CreatedAt.After(newestComment.CreatedAt) {
			newestComment = &comment
		}
	}

	var newestReview *PrReview
	for _, review := range pr.Reviews {
		if strings.Contains(review.User.Login, "[bot]") {
			continue
		}
		if newestReview == nil || review.SubmittedAt.After(newestReview.SubmittedAt) {
			newestReview = &review
		}
	}

	var newestCommit *PrCommit
	for _, commit := range pr.Commits {
		if newestCommit == nil || commit.Commit.Author.Date.After(newestCommit.Commit.Author.Date) {
			newestCommit = &commit
		}
	}

	actions := []Action{}

	actions = append(actions, Action{
		Type: "PR Opened",
		Time: pr.Pr.CreatedAt,
		User: pr.Pr.User.Login,
	})

	if newestComment != nil {
		actions = append(actions, Action{
			Type: "comment",
			Time: newestComment.CreatedAt,
			User: newestComment.User.Login,
		})
	}
	if newestReview != nil {
		actions = append(actions, Action{
			Type: "review",
			Time: newestReview.SubmittedAt,
			User: newestReview.User.Login,
		})
	}
	if newestCommit != nil {
		actions = append(actions, Action{
			Type: "commit",
			Time: newestCommit.Commit.Author.Date,
			// for now just assume the author is the only person commiting to the PR
			// github is weird with how it reports authors on PRs.
			User: pr.Pr.User.Login,
		})
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Time.After(actions[j].Time)
	})

	return actions
}

func sortPRs(prs []ProcessedPr, config *Config) []Section {
	var myPrsNeedAttention = Section{
		Name:         "My Prs that need attention",
		PullRequests: []ProcessedPr{},
	}
	var myPrsNoActionNeeded = Section{
		Name:         "My Prs no action needed",
		PullRequests: []ProcessedPr{},
	}
	var teamsPrsReadyForReview = Section{
		Name:         "Team Prs ready for review",
		PullRequests: []ProcessedPr{},
	}
	var teamPrsUnreviewedActions = Section{
		Name:         "Team PRs where Author has last action",
		PullRequests: []ProcessedPr{},
	}
	var teamPrsReviewedActions = Section{
		Name:         "Team PRs where someone else has last action",
		PullRequests: []ProcessedPr{},
	}

	for _, pr := range prs {
		if pr.RawPr.Pr.User.Login == config.MyGithubUser {
			if pr.Actions[0].User != config.MyGithubUser {
				myPrsNeedAttention.PullRequests = append(myPrsNeedAttention.PullRequests, pr)
			} else {
				myPrsNoActionNeeded.PullRequests = append(myPrsNoActionNeeded.PullRequests, pr)
			}
		} else {

			if !pr.ChecksFailed && pr.Actions[0].User == pr.RawPr.Pr.User.Login && !pr.ChangeRequested && !pr.Approved {
				teamsPrsReadyForReview.PullRequests = append(teamsPrsReadyForReview.PullRequests, pr)
			}

			if pr.Actions[0].User == pr.RawPr.Pr.User.Login {
				teamPrsUnreviewedActions.PullRequests = append(teamPrsUnreviewedActions.PullRequests, pr)
			} else {
				teamPrsReviewedActions.PullRequests = append(teamPrsReviewedActions.PullRequests, pr)
			}
		}
	}

	return []Section{myPrsNoActionNeeded, teamPrsReviewedActions, teamPrsUnreviewedActions, teamsPrsReadyForReview, myPrsNeedAttention}
}

func prHasChangesRequested(pr PullRequest) bool {
	for _, review := range pr.Reviews {
		if review.State == "CHANGE_REQUESTED" {
			return true
		}
	}
	return false
}

func prIsApproved(pr PullRequest) bool {
	for _, review := range pr.Reviews {
		if review.State == "APPROVED" {
			return true
		}
	}
	return false
}

func processPr(pr PullRequest) ProcessedPr {
	var actions = getActions(pr)
	var checksRunning = false
	var checksFailed = false
	for _, check := range pr.Checks.CheckRuns {
		if check.Status == "in_progress" {
			checksRunning = true
		}
		if check.Conclusion == "failure" {
			checksFailed = true
		}
	}
	return ProcessedPr{
		RawPr:              pr,
		Actions:            actions,
		Reviewed:           len(pr.Reviews) > 0,
		ChangeRequested:    prHasChangesRequested(pr),
		Approved:           prIsApproved(pr),
		ChecksStillRunning: checksRunning,
		ChecksFailed:       checksFailed,
	}
}

func printPrDetails(pr ProcessedPr) {
	fmt.Printf("PR:          %s%s %s| %s\n", RED, pr.RawPr.Pr.User.Login, RESET, pr.RawPr.Pr.Title)
	fmt.Printf("Url:         %s\n", pr.RawPr.Pr.Links.Html.Href)

	local, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading local time zone: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Actions:     ")
	for _, action := range pr.Actions {
		fmt.Printf("%s%s %sby %s%s%s at %s%s%s | ", GREEN, action.Type, RESET, RED, action.User, RESET, YELLOW, action.Time.In(local).Format(time.RFC822), RESET)
	}
	fmt.Printf("\n")

	var stateCharacter = fmt.Sprintf("%s%s", GREEN, RESET)
	if pr.RawPr.Pr.Draft {
		stateCharacter = fmt.Sprintf("")
	}
	var reviewedCharacter = ""
	if pr.Reviewed {
		reviewedCharacter = fmt.Sprintf("")
	}
	var changeRequestedCharacter = ""
	if pr.ChangeRequested {
		changeRequestedCharacter = fmt.Sprintf("%s%s", RED, RESET)
	}
	var approvedCharacter = ""
	if pr.Approved {
		approvedCharacter = fmt.Sprintf("%s%s", GREEN, RESET)
	}
	var runningCharacter = ""
	if pr.ChecksStillRunning {
		runningCharacter = fmt.Sprintf("%s%s", YELLOW, RESET)
	}
	var failedCharacter = ""
	if pr.ChecksFailed {
		failedCharacter = fmt.Sprintf("%s%s", RED, RESET)
	}
	fmt.Printf("Info:        %s  %s %s %s\n", stateCharacter, reviewedCharacter, changeRequestedCharacter, approvedCharacter)
	fmt.Printf("Checks:      %s %s\n", runningCharacter, failedCharacter)

	fmt.Printf("\n")
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	var searchUsers []string
	for _, user := range config.TeamUsers {
		searchUsers = append(searchUsers, fmt.Sprintf("author:%s", user))
	}

	searchParams := map[string]string{
		"q":        fmt.Sprintf("is:pr is:open org:%s author:@me %s", config.Organization, strings.Join(searchUsers, " ")),
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
	var processed []ProcessedPr
	for _, pr := range prs {
		processed = append(processed, processPr(pr))
	}
	var sections = sortPRs(processed, config)

	for _, section := range sections {
		fmt.Printf("#### %s ####\n", section.Name)
		for _, pr := range section.PullRequests {
			printPrDetails(pr)
		}
		fmt.Printf("\n")
	}
}
