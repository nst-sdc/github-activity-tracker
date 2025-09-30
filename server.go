package main

import (
    "encoding/json"
    "fmt"
    "gofr.dev/pkg/gofr"
    "net/http"
    "time"
)

func main() {

    InitDB()


    app := gofr.New()


    app.GET("/greet", func(ctx *gofr.Context) (any, error) {
        return "Hello World!", nil
    })

    func fetchPRsFromGitHub(username string, month string) ([]PR, error) {

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

        var prs []PR
        for _, item := range result.Items {
            prs = append(prs, PR{
                Title: item.Title,
                State: item.State,
                URL:   item.HtmlURL,
                
            })
        }
        return prs, nil
    }

    app.POST("/users", func(ctx *gofr.Context) (any, error) {
        var user User
        if err := ctx.Bind(&user); err != nil {
            return nil, err
        }
        DB.Create(&user)

        months := []string{"2025-09", "2025-10"}
        for _, monthName := range months {
            prs, err := fetchPRsFromGitHub(user.GithubUser, monthName)
            if err == nil {
                var month Month
                DB.Where("name = ?", monthName).FirstOrCreate(&month, Month{Name: monthName})
                for _, pr := range prs {
                    pr.UserID = user.ID
                    pr.MonthID = month.ID
                    DB.Create(&pr)
                }
            }
        }
        return user, nil
    })
    // Leaderboard endpoint: returns all users and their PR count for Sept/Oct
    app.GET("/leaderboard", func(ctx *gofr.Context) (any, error) {
        var users []User
        DB.Preload("PRs.Month").Find(&users)

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
        var users []User
        DB.Preload("PRs.Month").Find(&users)

        var dashboard []map[string]interface{}
        for _, user := range users {
            var prs []PR
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
            var user User
            if err := DB.Where("github_user = ?", username).First(&user).Error; err != nil {
                results = append(results, map[string]interface{}{
                    "username": username,
                    "prs":      nil,
                })
                continue
            }

            var month Month
            if err := DB.Where("name = ?", req.MonthName).First(&month).Error; err != nil {
                results = append(results, map[string]interface{}{
                    "username": username,
                    "prs":      nil,
                })
                continue
            }

            var prs []PR
     
            if err := DB.Preload("User").Preload("Org").Preload("Project").Preload("Month").
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
                    "id":       pr.ID,
                    "title":    pr.Title,
                    "url":      pr.URL,
                    "state":    pr.State,
                    "merged":   pr.Merged,
                    "created":  pr.CreatedAt,
                    "org":      pr.Org.Name,
                    "project":  pr.Project.Name,
                    "month":    pr.Month.Name,
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
