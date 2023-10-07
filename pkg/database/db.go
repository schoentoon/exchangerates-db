package database

import "gorm.io/gorm"

type CurrencyRate struct {
	Date         string  `gorm:"type:date;primaryKey"` // date in YYYY-MM-DD format
	Rate         float64 `gorm:"not null"`             // exchange rate as a floating point number
	FromCurrency string  `gorm:"not null;primaryKey"`  // three-letter ISO 4217 currency code for the base currency
	ToCurrency   string  `gorm:"not null;primaryKey"`  // three-letter ISO 4217 currency code for the target currency
}

func Init(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&CurrencyRate{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
