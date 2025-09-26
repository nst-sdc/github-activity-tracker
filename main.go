package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
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

type UserResponse struct {
	User         BasicUser         `json:"user"`
	Repos        []RepoInfo        `json:"repos"`
	Followers    []BasicUser       `json:"followers"`
	Following    []BasicUser       `json:"following"`
	PullRequests []PullRequestInfo `json:"pull_requests"`
}

func main() {
	client := newGitHubClientFromEnv()
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		handleUser(w, r, client)
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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

func handleUser(w http.ResponseWriter, r *http.Request, client *github.Client) {
	ctx := context.Background()
	username, err := extractUsername(r)
	if err != nil || username == "" {
		http.Error(w, "username is required (query param or JSON body)", http.StatusBadRequest)
		return
	}

	period := r.URL.Query().Get("period")
	var since time.Time
	switch period {
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "1m":
		since = time.Now().AddDate(0, -1, 0)
	default:
		since = time.Now().AddDate(0, -1, 0)
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

	prList, err := fetchUserPRs(ctx, client, username, repos, since)
	if err != nil {
		http.Error(w, "error fetching PRs: "+err.Error(), http.StatusBadGateway)
		return
	}

	resp := UserResponse{
		User:         mapUser(user),
		Repos:        mapRepos(repos),
		Followers:    mapBasicUsers(followers),
		Following:    mapBasicUsers(following),
		PullRequests: prList,
	}

	writeJSON(w, resp)
}

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

func fetchUserPRs(ctx context.Context, client *github.Client, username string, repos []*github.Repository, since time.Time) ([]PullRequestInfo, error) {
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
				if pr.CreatedAt.Before(since) && (pr.MergedAt == nil || pr.MergedAt.Before(since)) {
					continue
				}
				merged := false
				if pr.MergedAt != nil && pr.MergedAt.After(since) {
					merged = true
				}
				allPRs = append(allPRs, PullRequestInfo{
					RepoName:   *repo.FullName,
					Title:      *pr.Title,
					URL:        *pr.HTMLURL,
					State:      *pr.State,
					Merged:     merged,
					CreatedAt:  pr.CreatedAt.String(),
					ClosedAt:   timestampOrEmpty(pr.ClosedAt),
					MergedAt:   timestampOrEmpty(pr.MergedAt),
					BaseBranch: pr.Base.GetRef(),
					HeadBranch: pr.Head.GetRef(),
				})
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}
	return allPRs, nil
}

func timestampOrEmpty(t *github.Timestamp) string {
	if t == nil {
		return ""
	}
	return t.String()
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
	out := make([]RepoInfo, 0, len(repos))
	for _, r := range repos {
		if r == nil {
			continue
		}
		ri := RepoInfo{}
		if r.Name != nil {
			ri.Name = *r.Name
		}
		if r.FullName != nil {
			ri.FullName = *r.FullName
		}
		if r.HTMLURL != nil {
			ri.HTMLURL = *r.HTMLURL
		}
		if r.Description != nil {
			ri.Description = *r.Description
		}
		if r.Language != nil {
			ri.Language = *r.Language
		}
		if r.StargazersCount != nil {
			ri.Stars = *r.StargazersCount
		}
		if r.ForksCount != nil {
			ri.Forks = *r.ForksCount
		}
		if r.Private != nil {
			ri.Private = *r.Private
		}
		out = append(out, ri)
	}
	return out
}

func mapBasicUsers(users []*github.User) []BasicUser {
	out := make([]BasicUser, 0, len(users))
	for _, u := range users {
		out = append(out, mapUser(u))
	}
	return out
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, "error encoding json: "+err.Error(), http.StatusInternalServerError)
	}
}
