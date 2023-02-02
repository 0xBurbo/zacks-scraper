package earningscalendar

import (
	"fmt"
	"log"
	"os"

	"github.com/iamburbo/zacks-scraper/util"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type SplitsDataRow struct {
	Symbol      string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company     string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MarketCap   string `parquet:"name=marketCap, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Price       string `parquet:"name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	SplitFactor string `parquet:"name=splitFactor, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func parseSplitsData(rawData *earningsCalendarRawData) []*SplitsDataRow {
	rows := []*SplitsDataRow{}

	for _, entry := range rawData.Data {
		row := parseSplitsEntry(entry)
		rows = append(rows, row)
	}

	return rows
}

func parseSplitsEntry(entry dataEntry) *SplitsDataRow {

	symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, entry[0])
	company := util.FindStringInBetween(`<span title="`, `" >`, entry[1])
	marketCap := entry[2]
	price := entry[3]
	splitFactor := entry[4]

	return &SplitsDataRow{
		Symbol:      symbol,
		Company:     company,
		MarketCap:   marketCap,
		Price:       price,
		SplitFactor: splitFactor,
	}
}

func writeSplitsData(w *os.File, data []*SplitsDataRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(SplitsDataRow), 4)
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
