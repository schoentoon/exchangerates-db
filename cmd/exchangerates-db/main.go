package main

import (
	"flag"
	"sync"

	"github.com/schoentoon/exchangerates-db/pkg/database"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	configFile := flag.String("config", "config.yml", "What config file to use?")
	flag.Parse()

	if *configFile == "" {
		logrus.Fatal("No config file specified")
	}

	cfg, err := ReadConfig(*configFile)
	if err != nil {
		logrus.Fatal(err)
	}

	if cfg.Debug {
		logrus.SetReportCaller(true)
		logrus.SetLevel(logrus.DebugLevel)
	}

	var connector gorm.Dialector

	switch cfg.DB.Driver {
	case "sqlite":
		connector = sqlite.Open(cfg.DB.DSN)
	case "mysql":
		connector = mysql.Open(cfg.DB.DSN)
	case "postgres":
		connector = postgres.Open(cfg.DB.DSN)
	default:
		logrus.Fatalf("Invalid DB driver specified: %s", cfg.DB.Driver)
	}

	db, err := database.Init(connector, &gorm.Config{})
	if err != nil {
		logrus.Fatal(err)
	}

	importers := []*Importer{}

	for _, importerCfg := range cfg.Importers {
		importer, err := NewImporter(importerCfg.Driver, importerCfg.When)
		if err != nil {
			logrus.Fatal(err)
		}
		importers = append(importers, importer)
	}

	var wg sync.WaitGroup

	for _, importer := range importers {
		if !importer.When.Startup {
			continue
		}

		wg.Add(1)

		go func(importer *Importer, wg *sync.WaitGroup) {
			err := importer.ImportAll(db)
			if err != nil {
				logrus.Error(err)
			}
			wg.Done()
		}(importer, &wg)

	}

	wg.Wait()
}
