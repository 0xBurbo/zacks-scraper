package earningscalendar

import (
	"fmt"
	"log"
	"os"

	"github.com/iamburbo/zacks-scraper/util"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type EarningsDataRow struct {
	Symbol             string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company            string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MarketCap          string `parquet:"name=marketCap, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Time               string `parquet:"name=time, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Estimate           string `parquet:"name=estimate, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Reported           string `parquet:"name=reported, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Surprise           string `parquet:"name=surprise, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PercentSurp        string `parquet:"name=percentSurp, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PercentPriceChange string `parquet:"name=percentPriceChange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func parseEarningsData(rawData *earningsCalendarRawData) []*EarningsDataRow {
	rows := []*EarningsDataRow{}
	for _, entry := range rawData.Data {
		row := parseEarningsEntry(entry)
		rows = append(rows, row)
	}

	return rows
}

func parseEarningsEntry(entry dataEntry) *EarningsDataRow {

	symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, entry[0])
	company := util.FindStringInBetween(`<span title="`, `" >`, entry[1])
	marketCap := entry[2]
	time := entry[3]
	estimate := entry[4]
	reported := entry[5]

	var surprise string
	if entry[6] != "--" {
		surprise = util.FindStringInBetween(`">`, `</div>`, entry[6])
	} else {
		surprise = entry[6]
	}

	var percentSurp string
	if entry[7] != "--" {
		percentSurp = util.FindStringInBetween(`">`, `</div>`, entry[7])
	} else {
		percentSurp = entry[7]
	}

	percentPriceChange := util.FindStringInBetween(`">`, `</div>`, entry[8])

	return &EarningsDataRow{
		Symbol:             symbol,
		Company:            company,
		MarketCap:          marketCap,
		Time:               time,
		Estimate:           estimate,
		Reported:           reported,
		Surprise:           surprise,
		PercentSurp:        percentSurp,
		PercentPriceChange: percentPriceChange,
	}
}

func writeEarningsData(w *os.File, data []*EarningsDataRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(EarningsDataRow), 4)
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
