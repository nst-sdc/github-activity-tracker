package main

import (
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
    dsn := "host=localhost user=postgres password=Vipyadav@2005 dbname=githubdata port=5432 sslmode=disable TimeZone=Asia/Kolkata"


    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    log.Println("Connected to PostgreSQL!")

    // Example model auto-migration
    DB.AutoMigrate(&User{})
}

// Example model
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string
    Email string `gorm:"unique"`
}
