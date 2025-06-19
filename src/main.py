import json
from client import fetch_gh
from models import Comment, PrCommit, PullRequest, PrDetail

search_params = {"q": "is:pr is:open org:MercuryTechnologies author:@me", "per_page": 10}  # Limit to 10 results for this example

response = fetch_gh(
    "/search/issues", params=search_params
)
assert isinstance(response, dict)

search_results = response.get("items", [])
if not search_results:
    exit(0)

pr_api_url = search_results[0].get("pull_request", {}).get("url")

if not pr_api_url:
    exit(1)

pr_details = fetch_gh(pr_api_url)

pr_detail = PrDetail.from_dict(pr_details)

comments_response = fetch_gh(pr_detail.comments_url)
assert isinstance(comments_response, list)
comments = Comment.schema().load(comments_response, many=True)

commits_response = fetch_gh(pr_detail.commits_url)
assert isinstance(commits_response, list)
commits = PrCommit.schema().load(commits_response, many=True)

full_pr = PullRequest(pr=pr_detail, comments=comments, commits=commits)
print(full_pr)
