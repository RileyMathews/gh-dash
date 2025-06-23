package main

import (
    "encoding/json"
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
        fmt.Fprintln(os.Stderr, "GitHub token not found. Please set the GITHUB_TOKEN environment variable.")
        os.Exit(1)
    }
    return token
}

// getHeaders prepares the HTTP headers for all requests.
func getHeaders() http.Header {
    headers := http.Header{}
    headers.Set("Accept", "application/vnd.github.v3+json")
    headers.Set("Authorization", "Bearer "+getToken())
    return headers
}

// fetchGH fetches the given path (or full URL) from GitHub, applies any query params,
// and unmarshals the JSON response directly into `out`.
// Returns a non-nil error on network, status-code, or JSON errors.
func fetchGH(path string, params map[string]string, out interface{}) error {
    var fullURL string
    if strings.HasPrefix(path, "http") {
        fullURL = path
    } else {
        fullURL = API_BASE_URL + path
    }

    // Add query parameters if provided
    if len(params) > 0 {
        u, err := url.Parse(fullURL)
        if err != nil {
            return fmt.Errorf("invalid URL %q: %w", fullURL, err)
        }
        q := u.Query()
        for k, v := range params {
            q.Set(k, v)
        }
        u.RawQuery = q.Encode()
        fullURL = u.String()
    }

    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    req.Header = getHeaders()

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("request error for %q: %w", fullURL, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        category := "Client Error"
        if resp.StatusCode >= 500 {
            category = "Server Error"
        }
        return fmt.Errorf("HTTP %d %s for URL %s", resp.StatusCode, category, fullURL)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("reading response body: %w", err)
    }

    if err := json.Unmarshal(body, out); err != nil {
        return fmt.Errorf("decoding JSON response: %w", err)
    }
    return nil
}

