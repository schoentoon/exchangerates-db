package main

import (
	"log"

	"github.com/schoentoon/exchangerates-db/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	_, err := database.Init(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}
