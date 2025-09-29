package main

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/google/go-github/v53/github"
    "golang.org/x/oauth2"
)

type RepoInfo struct {
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    HTMLURL     string `json:"html_url"`
    Description string `json:"description,omitempty"`
    Language    string `json:"language,omitempty"`
    Stars       int    `json:"stargazers_count"`
    Forks       int    `json:"forks_count"`
    Private     bool   `json:"private"`
}

type BasicUser struct {
    Login       string `json:"login,omitempty"`
    ID          int64  `json:"id,omitempty"`
    Name        string `json:"name,omitempty"`
    Company     string `json:"company,omitempty"`
    Blog        string `json:"blog,omitempty"`
    Location    string `json:"location,omitempty"`
    Email       string `json:"email,omitempty"`
    Bio         string `json:"bio,omitempty"`
    Followers   int    `json:"followers,omitempty"`
    Following   int    `json:"following,omitempty"`
    PublicRepos int    `json:"public_repos,omitempty"`
    AvatarURL   string `json:"avatar_url,omitempty"`
    HTMLURL     string `json:"html_url,omitempty"`
}

type PullRequestInfo struct {
    RepoName   string `json:"repo_name"`
    Title      string `json:"title"`
    URL        string `json:"url"`
    State      string `json:"state"`
    Merged     bool   `json:"merged"`
    CreatedAt  string `json:"created_at"`
    ClosedAt   string `json:"closed_at,omitempty"`
    MergedAt   string `json:"merged_at,omitempty"`
    BaseBranch string `json:"base_branch,omitempty"`
    HeadBranch string `json:"head_branch,omitempty"`
}

// New type for commits info
type CommitInfo struct {
    RepoName    string `json:"repo_name"`
    CommitSHA   string `json:"commit_sha"`
    Message     string `json:"message"`
    URL         string `json:"url"`
    AuthorName  string `json:"author_name"`
    AuthorEmail string `json:"author_email"`
    Date        string `json:"date"`
}

type UserResponse struct {
    User      BasicUser   `json:"user"`
    Repos     []RepoInfo  `json:"repos"`
    Followers []BasicUser `json:"followers"`
    Following []BasicUser `json:"following"`
}

func main() {
    client := newGitHubClientFromEnv()

    http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
        handleUser(w, r, client)
    })

    http.HandleFunc("/pulls", func(w http.ResponseWriter, r *http.Request) {
        username := r.URL.Query().Get("username")
        if username == "" {
            http.Error(w, "username is required", http.StatusBadRequest)
            return
        }

        period := r.URL.Query().Get("period") // e.g., "24h", "7d", "30d"
        var since time.Time
        if period != "" {
            d, err := parsePeriod(period)
            if err != nil {
                http.Error(w, "invalid period format: "+err.Error(), http.StatusBadRequest)
                return
            }
            since = time.Now().UTC().Add(-d)
        }

        pulls, err := fetchPullRequests(client, username, since)
        if err != nil {
            http.Error(w, "error fetching pull requests: "+err.Error(), http.StatusInternalServerError)
            return
        }
        if pulls == nil {
            pulls = []PullRequestInfo{}
        }

        commits, err := fetchUserCommits(client, username, since)
        if err != nil {
            http.Error(w, "error fetching commits: "+err.Error(), http.StatusInternalServerError)
            return
        }
        if commits == nil {
            commits = []CommitInfo{}
        }

        writeJSON(w, map[string]interface{}{
            "pull_requests": pulls,
            "commits":       commits,
        })
    })

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- GitHub client ---
func newGitHubClientFromEnv() *github.Client {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        log.Println("GITHUB_TOKEN not set — using unauthenticated client (lower rate limits)")
        return github.NewClient(nil)
    }
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(ctx, ts)
    return github.NewClient(tc)
}

// --- User handler ---
func handleUser(w http.ResponseWriter, r *http.Request, client *github.Client) {
    ctx := context.Background()
    username, err := extractUsername(r)
    if err != nil || username == "" {
        http.Error(w, "username is required (query param or JSON body)", http.StatusBadRequest)
        return
    }

    user, _, err := client.Users.Get(ctx, username)
    if err != nil {
        http.Error(w, "error fetching user: "+err.Error(), http.StatusBadGateway)
        return
    }

    repos, err := fetchAllRepos(ctx, client, username)
    if err != nil {
        http.Error(w, "error fetching repos: "+err.Error(), http.StatusBadGateway)
        return
    }

    followers, err := fetchUsersList(ctx, client, username, "followers")
    if err != nil {
        http.Error(w, "error fetching followers: "+err.Error(), http.StatusBadGateway)
        return
    }
    following, err := fetchUsersList(ctx, client, username, "following")
    if err != nil {
        http.Error(w, "error fetching following: "+err.Error(), http.StatusBadGateway)
        return
    }

    resp := UserResponse{
        User:      mapUser(user),
        Repos:     mapRepos(repos),
        Followers: mapBasicUsers(followers),
        Following: mapBasicUsers(following),
    }

    writeJSON(w, resp)
}

// --- Extract username from query or body ---
func extractUsername(r *http.Request) (string, error) {
    q := r.URL.Query().Get("username")
    if q != "" {
        return q, nil
    }
    if r.Method == http.MethodPost {
        defer r.Body.Close()
        body, _ := io.ReadAll(r.Body)
        if len(body) == 0 {
            return "", nil
        }
        var payload map[string]string
        if err := json.Unmarshal(body, &payload); err != nil {
            return "", err
        }
        if u, ok := payload["username"]; ok {
            return u, nil
        }
    }
    return "", nil
}

// --- Fetch all repos ---
func fetchAllRepos(ctx context.Context, client *github.Client, username string) ([]*github.Repository, error) {
    opts := &github.RepositoryListOptions{
        Type:      "owner",
        Sort:      "updated",
        Direction: "desc",
        ListOptions: github.ListOptions{
            PerPage: 100,
        },
    }
    var all []*github.Repository
    for {
        repos, resp, err := client.Repositories.List(ctx, username, opts)
        if err != nil {
            return nil, err
        }
        all = append(all, repos...)
        if resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }
    return all, nil
}

// --- Fetch followers or following ---
func fetchUsersList(ctx context.Context, client *github.Client, username, kind string) ([]*github.User, error) {
    opts := &github.ListOptions{PerPage: 100}
    var all []*github.User
    for {
        var users []*github.User
        var resp *github.Response
        var err error
        if kind == "followers" {
            users, resp, err = client.Users.ListFollowers(ctx, username, opts)
        } else if kind == "following" {
            users, resp, err = client.Users.ListFollowing(ctx, username, opts)
        } else {
            return nil, errors.New("unknown kind")
        }
        if err != nil {
            return nil, err
        }
        all = append(all, users...)
        if resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }
    return all, nil
}

// --- Fetch pull requests ---
func fetchPullRequests(client *github.Client, username string, since time.Time) ([]PullRequestInfo, error) {
    ctx := context.Background()
    repos, err := fetchAllRepos(ctx, client, username)
    if err != nil {
        return nil, err
    }

    var allPRs []PullRequestInfo
    for _, repo := range repos {
        if repo.Name == nil || repo.Owner == nil || repo.Owner.Login == nil {
            continue
        }
        opts := &github.PullRequestListOptions{
            State: "all",
            ListOptions: github.ListOptions{
                PerPage: 50,
            },
        }
        for {
            prs, resp, err := client.PullRequests.List(ctx, *repo.Owner.Login, *repo.Name, opts)
            if err != nil {
                return nil, err
            }
            for _, pr := range prs {
                if since.IsZero() || pr.CreatedAt.UTC().After(since) {
                    merged := pr.MergedAt != nil && !pr.MergedAt.IsZero()
                    allPRs = append(allPRs, PullRequestInfo{
                        RepoName:   stringOrNil(repo.FullName),
                        Title:      stringOrNil(pr.Title),
                        URL:        stringOrNil(pr.HTMLURL),
                        State:      stringOrNil(pr.State),
                        Merged:     merged,
                        CreatedAt:  pr.CreatedAt.Format(time.RFC3339),
                        ClosedAt:   timestampOrEmpty(pr.ClosedAt),
                        MergedAt:   timestampOrEmpty(pr.MergedAt),
                        BaseBranch: pr.Base.GetRef(),
                        HeadBranch: pr.Head.GetRef(),
                    })
                }
            }
            if resp.NextPage == 0 {
                break
            }
            opts.Page = resp.NextPage
        }
    }

    if allPRs == nil {
        allPRs = []PullRequestInfo{}
    }

    return allPRs, nil
}

// --- Fetch commits authored by user ---
func fetchUserCommits(client *github.Client, username string, since time.Time) ([]CommitInfo, error) {
    ctx := context.Background()
    repos, err := fetchAllRepos(ctx, client, username)
    if err != nil {
        return nil, err
    }

    var allCommits []CommitInfo
    for _, repo := range repos {
        if repo.Name == nil || repo.Owner == nil || repo.Owner.Login == nil {
            continue
        }

        opts := &github.CommitsListOptions{
            Author: username,
            ListOptions: github.ListOptions{
                PerPage: 50,
            },
        }
        if !since.IsZero() {
            opts.Since = since
        }

        for {
            commits, resp, err := client.Repositories.ListCommits(ctx, *repo.Owner.Login, *repo.Name, opts)
            if err != nil {
                return nil, err
            }

            for _, c := range commits {
                if c.Commit == nil || c.SHA == nil || c.HTMLURL == nil {
                    continue
                }
                authorName := ""
                authorEmail := ""
                commitDate := ""
                if c.Commit.Author != nil {
                    authorName = c.Commit.Author.Name
                    authorEmail = c.Commit.Author.Email
                    commitDate = c.Commit.Author.Date.Format(time.RFC3339)
                }

                allCommits = append(allCommits, CommitInfo{
                    RepoName:    stringOrNil(repo.FullName),
                    CommitSHA:   *c.SHA,
                    Message:     stringOrNil(c.Commit.Message),
                    URL:         *c.HTMLURL,
                    AuthorName:  authorName,
                    AuthorEmail: authorEmail,
                    Date:        commitDate,
                })
            }
            if resp.NextPage == 0 {
                break
            }
            opts.Page = resp.NextPage
        }
    }

    if allCommits == nil {
        allCommits = []CommitInfo{}
    }

    return allCommits, nil
}

// --- Utilities ---
func timestampOrEmpty(t *github.Timestamp) string {
    if t == nil || t.Time.IsZero() {
        return ""
    }
    return t.Time.Format(time.RFC3339)
}

func mapUser(u *github.User) BasicUser {
    b := BasicUser{}
    if u == nil {
        return b
    }
    if u.Login != nil {
        b.Login = *u.Login
    }
    if u.ID != nil {
        b.ID = *u.ID
    }
    if u.Name != nil {
        b.Name = *u.Name
    }
    if u.Company != nil {
        b.Company = *u.Company
    }
    if u.Blog != nil {
        b.Blog = *u.Blog
    }
    if u.Location != nil {
        b.Location = *u.Location
    }
    if u.Email != nil {
        b.Email = *u.Email
    }
    if u.Bio != nil {
        b.Bio = *u.Bio
    }
    if u.Followers != nil {
        b.Followers = *u.Followers
    }
    if u.Following != nil {
        b.Following = *u.Following
    }
    if u.PublicRepos != nil {
        b.PublicRepos = *u.PublicRepos
    }
    if u.AvatarURL != nil {
        b.AvatarURL = *u.AvatarURL
    }
    if u.HTMLURL != nil {
        b.HTMLURL = *u.HTMLURL
    }
    return b
}

func mapRepos(repos []*github.Repository) []RepoInfo {
    var out []RepoInfo
    for _, r := range repos {
        out = append(out, RepoInfo{
            Name:        stringOrNil(r.Name),
            FullName:    stringOrNil(r.FullName),
            HTMLURL:     stringOrNil(r.HTMLURL),
            Description: stringOrNil(r.Description),
            Language:    stringOrNil(r.Language),
            Stars:       intOrNil(r.StargazersCount),
            Forks:       intOrNil(r.ForksCount),
            Private:     boolOrNil(r.Private),
        })
    }
    return out
}

func mapBasicUsers(users []*github.User) []BasicUser {
    var out []BasicUser
    for _, u := range users {
        out = append(out, mapUser(u))
    }
    return out
}

func stringOrNil(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func intOrNil(i *int) int {
    if i == nil {
        return 0
    }
    return *i
}

func boolOrNil(b *bool) bool {
    if b == nil {
        return false
    }
    return *b
}

func writeJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
}

// --- Parse period string like "24h", "7d", "1w", "1m" ---
func parsePeriod(period string) (time.Duration, error) {
    if len(period) < 2 {
        return 0, errors.New("too short")
    }
    unit := period[len(period)-1]
    numStr := period[:len(period)-1]
    num, err := strconv.ParseFloat(numStr, 64)
    if err != nil {
        return 0, err
    }

    switch unit {
    case 'h':
        return time.Duration(num * float64(time.Hour)), nil
    case 'd':
        return time.Duration(num * float64(24*time.Hour)), nil
    case 'w':
        return time.Duration(num * float64(7*24*time.Hour)), nil
    case 'm':
        return time.Duration(num * float64(30*24*time.Hour)), nil
    default:
        return 0, errors.New("unknown time unit")
    }
}
