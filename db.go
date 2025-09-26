package main

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)
func main() {
    //Create a new Postgresql database connection
    dsn := "host=<your_host> user=<your_user> \
    password=<your_password> dbname=<your_dbname> port=<your_port>"

    // Open a connection to the database
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database: " + err.Error())
    }
}

