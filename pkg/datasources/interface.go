package datasources

import "github.com/schoentoon/exchangerates-db/pkg/database"

type Datasource interface {
	ImportAll(chan<- database.CurrencyRate) error
	ImportLatest(chan<- database.CurrencyRate) error
}
