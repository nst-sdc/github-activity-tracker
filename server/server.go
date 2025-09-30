package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gofr.dev/pkg/gofr"

	"github-activity-tracker/DB"
	"github-activity-tracker/models"
)

func fetchPRsFromGitHub(username string, month string) ([]models.PR, error) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:pr+created:%s", username, month)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			State   string `json:"state"`
			HtmlURL string `json:"html_url"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var prs []models.PR
	for _, item := range result.Items {
		prs = append(prs, models.PR{
			Title:  item.Title,
			Status: item.State,
			URL:    item.HtmlURL,
			Merged: item.State == "closed", // Approximate - would need more API calls to determine if actually merged
		})
	}
	return prs, nil
}

func main() {
	DB.InitDB()
	db := DB.GetDB()

	app := gofr.New()

	app.GET("/greet", func(ctx *gofr.Context) (any, error) {
		return "Hello World!", nil
	})

	app.POST("/users", func(ctx *gofr.Context) (any, error) {
		var user models.User
		if err := ctx.Bind(&user); err != nil {
			return nil, err
		}
		db.Create(&user)

		months := []string{"2025-09", "2025-10"}
		for _, monthName := range months {
			prs, err := fetchPRsFromGitHub(user.GithubUser, monthName)
			if err == nil {
				var month models.Month
				db.Where("name = ?", monthName).FirstOrCreate(&month, models.Month{Name: monthName})
				for _, pr := range prs {
					pr.UserID = user.ID
					pr.MonthID = month.ID
					db.Create(&pr)
				}
			}
		}
		return user, nil
	})

	app.GET("/leaderboard", func(ctx *gofr.Context) (any, error) {
		var users []models.User
		db.Preload("PRs.Month").Find(&users)

		var leaderboard []map[string]interface{}
		for _, user := range users {
			prCount := 0
			for _, pr := range user.PRs {
				if pr.Month.Name == "2025-09" || pr.Month.Name == "2025-10" {
					prCount++
				}
			}
			leaderboard = append(leaderboard, map[string]interface{}{
				"name":        user.Name,
				"github_user": user.GithubUser,
				"pr_count":    prCount,
			})
		}
		return leaderboard, nil
	})

	app.GET("/admin-dashboard", func(ctx *gofr.Context) (any, error) {
		var users []models.User
		db.Preload("PRs.Month").Find(&users)

		var dashboard []map[string]interface{}
		for _, user := range users {
			var prs []models.PR
			for _, pr := range user.PRs {
				if pr.Month.Name == "2025-09" || pr.Month.Name == "2025-10" {
					prs = append(prs, pr)
				}
			}
			dashboard = append(dashboard, map[string]interface{}{
				"name":        user.Name,
				"github_user": user.GithubUser,
				"prs":         prs,
			})
		}
		return dashboard, nil
	})

	app.POST("/track-prs", func(ctx *gofr.Context) (any, error) {
		var req struct {
			Usernames []string `json:"usernames"`
			MonthName string   `json:"month_name"`
		}
		if err := ctx.Bind(&req); err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, username := range req.Usernames {
			var user models.User
			if err := db.Where("github_user = ?", username).First(&user).Error; err != nil {
				results = append(results, map[string]interface{}{
					"username": username,
					"prs":      nil,
				})
				continue
			}

			var month models.Month
			if err := db.Where("name = ?", req.MonthName).First(&month).Error; err != nil {
				results = append(results, map[string]interface{}{
					"username": username,
					"prs":      nil,
				})
				continue
			}

			var prs []models.PR
			if err := db.Preload("User").Preload("Org").Preload("Project").Preload("Month").
				Where("user_id = ? AND month_id = ?", user.ID, month.ID).Find(&prs).Error; err != nil {
				results = append(results, map[string]interface{}{
					"username": username,
					"prs":      nil,
				})
				continue
			}

			var prList []map[string]interface{}
			for _, pr := range prs {
				prList = append(prList, map[string]interface{}{
					"id":      pr.ID,
					"title":   pr.Title,
					"url":     pr.URL,
					"status":  pr.Status,
					"merged":  pr.Merged,
					"created": pr.CreatedAt,
					"org":     pr.Org.Name,
					"project": pr.Project.Name,
					"month":   pr.Month.Name,
				})
			}

			results = append(results, map[string]interface{}{
				"username": username,
				"prs":      prList,
			})
		}
		return results, nil
	})

	app.Run()
}
