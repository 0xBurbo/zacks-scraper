package earningscalendar

import (
	"fmt"
	"log"
	"os"

	"github.com/iamburbo/zacks-scraper/util"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type DividendsDataRow struct {
	Symbol       string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company      string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MarketCap    string `parquet:"name=marketCap, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Amount       string `parquet:"name=amount, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Yield        string `parquet:"name=yield, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	ExDivDate    string `parquet:"name=exDivDate, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CurrentPrice string `parquet:"name=currentPrice, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PayableDate  string `parquet:"name=payableDate, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func parseDividendsData(rawData *earningsCalendarRawData) []*DividendsDataRow {
	rows := []*DividendsDataRow{}
	for _, entry := range rawData.Data {
		row := parseDividendsEntry(entry)
		rows = append(rows, row)
	}

	return rows
}

func parseDividendsEntry(entry dataEntry) *DividendsDataRow {

	symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, entry[0])
	company := util.FindStringInBetween(`<span title="`, `" >`, entry[1])
	marketCap := entry[2]
	amount := entry[3]
	yield := entry[4]
	exDivDate := entry[5]
	currentPrice := entry[6]
	payableDate := entry[7]

	return &DividendsDataRow{
		Symbol:       symbol,
		Company:      company,
		MarketCap:    marketCap,
		Amount:       amount,
		Yield:        yield,
		ExDivDate:    exDivDate,
		CurrentPrice: currentPrice,
		PayableDate:  payableDate,
	}
}

func writeDividendsData(w *os.File, data []*DividendsDataRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(DividendsDataRow), 4)
	if err != nil {
		return fmt.Errorf("can't create parquet writer: %e", err)
	}

	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	for _, row := range data {
		if err = pw.Write(row); err != nil {
			log.Println("Write error", err)
		}
	}

	if err = pw.WriteStop(); err != nil {
		log.Println("WriteStop error", err)
	}

	return nil
}
