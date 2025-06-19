package main

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strings"
)

// Base URL for GitHub API v3
const API_BASE_URL = "https://api.github.com"

// getToken reads the GitHub token from the environment.
func getToken() string {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        // If no token is found, exit with an error message.
        fmt.Fprintln(os.Stderr, "GitHub token not found. Please set the GITHUB_TOKEN environment variable.")
        os.Exit(1)
    }
    return token
}

// getHeaders prepares the HTTP headers for all requests (including authentication).
func getHeaders() http.Header {
    headers := http.Header{}
    headers.Set("Accept", "application/vnd.github.v3+json")
    headers.Set("Authorization", "Bearer "+getToken())
    return headers
}

// fetchGH performs a GET request to the GitHub API (v3).
// The `path` can be a full URL or a path (e.g., "/search/issues") relative to API_BASE_URL.
// The `params` map contains query parameters to include in the request (or nil if none).
// On error (network issue or non-2xx status), it prints an error message and exits.
func fetchGH(path string, params map[string]string) []byte {
    // Determine the full URL to request
    var fullURL string
    if strings.HasPrefix(path, "http") {
        fullURL = path
    } else {
        fullURL = API_BASE_URL + path
    }

    // Append query parameters if provided
    if params != nil && len(params) > 0 {
        parsedURL, err := url.Parse(fullURL)
        if err != nil {
            fmt.Printf("CLIENT: Error fetching data from %s: %v\n", fullURL, err)
            os.Exit(1)
        }
        query := parsedURL.Query()
        for key, value := range params {
            query.Set(key, value)
        }
        parsedURL.RawQuery = query.Encode()
        fullURL = parsedURL.String()
    }

    // Create the HTTP GET request with appropriate headers
    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        fmt.Printf("CLIENT: Error fetching data from %s: %v\n", fullURL, err)
        os.Exit(1)
    }
    req.Header = getHeaders()

    // Execute the request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("CLIENT: Error fetching data from %s: %v\n", fullURL, err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    // Check for HTTP status errors (4xx or 5xx responses)
    if resp.StatusCode >= 400 {
        // Construct an error message similar to Python requests' HTTPError
        category := "Client Error"
        if resp.StatusCode >= 500 {
            category = "Server Error"
        }
        statusReason := http.StatusText(resp.StatusCode)
        errMsg := fmt.Sprintf("%d %s: %s for url: %s", resp.StatusCode, category, statusReason, fullURL)
        fmt.Printf("CLIENT: Error fetching data from %s: %s\n", fullURL, errMsg)
        os.Exit(1)
    }

    // Read the response body fully
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("CLIENT: Error reading response from %s: %v\n", fullURL, err)
        os.Exit(1)
    }
    return body
}

