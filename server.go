package main

import (
    "gofr.dev/pkg/gofr"
)

func main() {
    // Initialize database
    InitDB()

    // Initialize Gofr app
    app := gofr.New()

    // Routes
    app.GET("/greet", func(ctx *gofr.Context) (any, error) {
        return "Hello World!", nil
    })

    // Example route to add user
    app.POST("/users", func(ctx *gofr.Context) (any, error) {
        var user User
        if err := ctx.Bind(&user); err != nil {
            return nil, err
        }
        DB.Create(&user)
        return user, nil
    })

    app.Run()
}
