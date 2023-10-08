package main

import (
	"log"

	"github.com/schoentoon/exchangerates-db/pkg/database"
	"github.com/schoentoon/exchangerates-db/pkg/datasources/ecb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	db, err := database.Init(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan database.CurrencyRate, 16)

	go func() {
		source := &ecb.Datasource{}
		err = source.Import(ch)
		if err != nil {
			log.Fatal(err)
		}
		close(ch)
	}()

	db.Transaction(func(tx *gorm.DB) error {
		// we buffer up to this amount of currency rates to insert them all at once
		// we do however allocate the entire buffer right away so we won't have to
		// constantly resize the array for it while building up our buffer etc
		BUFFER_SIZE := 10000
		buffer := make([]database.CurrencyRate, BUFFER_SIZE)

		pos := 0
		for data := range ch {
			buffer[pos] = data

			pos++

			// our buffer is full, so we insert it and we reset our position back to the start
			// no need to clear the buffer, we'll just write over the already inserted rates
			if pos == (BUFFER_SIZE - 1) {
				db.Clauses(clause.OnConflict{DoNothing: true}).Create(buffer)
				pos = 0
			}
		}

		db.Clauses(clause.OnConflict{DoNothing: true}).Create(buffer)
		return nil
	})
}
