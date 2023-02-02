package earningscalendar

import (
	"fmt"
	"log"
	"os"

	"github.com/iamburbo/zacks-scraper/util"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type RevisionsDataRow struct {
	Symbol       string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company      string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MarketCap    string `parquet:"name=marketCap, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Period       string `parquet:"name=period, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PeriodEnd    string `parquet:"name=periodEnd, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Old          string `parquet:"name=old, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	New          string `parquet:"name=new, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	EstChange    string `parquet:"name=estChange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Cons         string `parquet:"name=cons, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	NewEstVsCons string `parquet:"name=newEstVsCons, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func parseRevisionsData(rawData *earningsCalendarRawData) []*RevisionsDataRow {
	rows := []*RevisionsDataRow{}

	for _, entry := range rawData.Data {
		row := parseRevisionsEntry(entry)
		rows = append(rows, row)
	}

	return rows
}

func parseRevisionsEntry(entry dataEntry) *RevisionsDataRow {

	symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, entry[0])
	company := util.FindStringInBetween(`<span title="`, `" >`, entry[1])
	marketCap := entry[2]
	period := entry[3]
	periodEnd := entry[4]
	old := entry[5]
	new := entry[6]

	var estChange string
	if entry[7] != "NA" {
		estChange = util.FindStringInBetween(`">`, `</div>`, entry[7])
	} else {
		estChange = entry[7]
	}

	cons := entry[8]

	var newEstVsCons string
	if entry[9] != "NA" {
		newEstVsCons = util.FindStringInBetween(`">`, `</div>`, entry[9])
	} else {
		newEstVsCons = entry[9]
	}

	return &RevisionsDataRow{
		Symbol:       symbol,
		Company:      company,
		MarketCap:    marketCap,
		Period:       period,
		PeriodEnd:    periodEnd,
		Old:          old,
		New:          new,
		EstChange:    estChange,
		Cons:         cons,
		NewEstVsCons: newEstVsCons,
	}
}

func writeRevisionsData(w *os.File, data []*RevisionsDataRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(RevisionsDataRow), 4)
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
