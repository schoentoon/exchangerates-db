package main

import (
	"fmt"

	"github.com/schoentoon/exchangerates-db/pkg/database"
	"github.com/schoentoon/exchangerates-db/pkg/datasources"
	"github.com/schoentoon/exchangerates-db/pkg/datasources/ecb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Importer struct {
	driver datasources.Datasource
	name   string

	When WhenConfig
}

type WhenConfig struct {
	Startup bool `yaml:"startup"`
}

func NewImporter(driver string, when WhenConfig) (*Importer, error) {
	out := &Importer{
		name: driver,
		When: when,
	}
	switch driver {
	case "ecb":
		out.driver = &ecb.Datasource{}
	default:
		return nil, fmt.Errorf("No such driver: %s", driver)
	}

	return out, nil
}

func (i *Importer) ImportAll(db *gorm.DB) error {
	return i.importWrapper(db, func(ch chan<- database.CurrencyRate) error {
		return i.driver.ImportAll(ch)
	})
}

func (i *Importer) ImportLatest(db *gorm.DB) error {
	return i.importWrapper(db, func(ch chan<- database.CurrencyRate) error {
		return i.driver.ImportLatest(ch)
	})
}

func (i *Importer) importWrapper(db *gorm.DB, f func(ch chan<- database.CurrencyRate) error) error {
	ch := make(chan database.CurrencyRate, 16)
	errCh := make(chan error, 1)

	go func() {
		err := f(ch)
		if err != nil {
			errCh <- err
		}
		close(ch)
	}()

	err := db.Transaction(func(tx *gorm.DB) error {
		// we buffer up to this amount of currency rates to insert them all at once
		// we do however allocate the entire buffer right away so we won't have to
		// constantly resize the array for it while building up our buffer etc
		BUFFER_SIZE := 1000
		buffer := make([]database.CurrencyRate, BUFFER_SIZE)

		pos := 0
		for data := range ch {
			data.Source = i.name
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
	if err != nil {
		return err
	}

	select {
	case err = <-errCh:
		return err
	default:
		return nil
	}
}
