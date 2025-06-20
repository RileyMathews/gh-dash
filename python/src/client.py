# github_client.py

import os
from typing import Any, List
import requests

# --- Configuration and Initialization ---

# The base URL for all GitHub API v3 calls.
API_BASE_URL = "https://api.github.com"

def get_token():

# Read the token from an environment variable for security.
    token = os.getenv("GITHUB_TOKEN")

# Check for the token at module load time.
    if not token:
        raise ValueError(
            "GitHub token not found. Please set the GITHUB_TOKEN environment variable."
        )
    return token

# Prepare the headers that will be used for every request.
# This ensures we are always authenticated.
def get_headers():
    return {
        "Accept": "application/vnd.github.v3+json",
        "Authorization": f"Bearer {get_token()}",
    }


def fetch_gh(path: str, params: dict | None = None) -> dict | List[Any]:
    if path.startswith("http"):
        full_url = path
    else:
        full_url = f"{API_BASE_URL}{path}"

    try:
        response = requests.get(full_url, headers=get_headers(), params=params)
        # Raise an exception for bad status codes (4xx or 5xx)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"CLIENT: Error fetching data from {full_url}: {e}")
        exit(1)
