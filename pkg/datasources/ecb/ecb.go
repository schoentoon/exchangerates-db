package ecb

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/schoentoon/exchangerates-db/pkg/database"
)

type Datasource struct {
}

func (d *Datasource) ImportAll(ch chan<- database.CurrencyRate) error {
	return d.importFromUrl(ch, "http://www.ecb.europa.eu/stats/eurofxref/eurofxref-hist.zip", "eurofxref-hist.csv")
}

func (d *Datasource) ImportLatest(ch chan<- database.CurrencyRate) error {
	return d.importFromUrl(ch, "http://www.ecb.europa.eu/stats/eurofxref/eurofxref.zip", "eurofxref.csv")
}

func (d *Datasource) importFromUrl(ch chan<- database.CurrencyRate, url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	buf := bytes.NewReader(b)
	unzipper, err := zip.NewReader(buf, int64(len(b)))
	if err != nil {
		return err
	}

	f, err := unzipper.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return d.handleZipfile(f, ch)
}

func (d *Datasource) handleZipfile(f fs.File, ch chan<- database.CurrencyRate) error {
	reader := csv.NewReader(f)

	header, err := reader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		date := record[0]

		for i, field := range record[1:] {
			rate, err := strconv.ParseFloat(field, 64)
			if err != nil {
				continue
			}

			ch <- database.CurrencyRate{
				Date:         date,
				FromCurrency: "EUR",
				ToCurrency:   header[i],
				Rate:         rate,
			}
		}
	}

	return nil
}
