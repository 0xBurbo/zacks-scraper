package earningscalendar

import (
	"fmt"
	"log"
	"os"

	"github.com/iamburbo/zacks-scraper/util"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type GuidanceDataRow struct {
	Symbol             string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company            string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MarketCap          string `parquet:"name=marketCap, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Period             string `parquet:"name=period, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PeriodEnd          string `parquet:"name=periodEnd, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	GuidRange          string `parquet:"name=guidRange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MidGuid            string `parquet:"name=midGuid, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Cons               string `parquet:"name=cons, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PercentToHighPoint string `parquet:"name=percentToHighPoint, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func parseGuidanceData(rawData *earningsCalendarRawData) []*GuidanceDataRow {
	rows := []*GuidanceDataRow{}
	for _, entry := range rawData.Data {
		row := parseGuidanceEntry(entry)
		rows = append(rows, row)
	}

	return rows
}

func parseGuidanceEntry(entry dataEntry) *GuidanceDataRow {

	symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, entry[0])
	company := util.FindStringInBetween(`<span title="`, `" >`, entry[1])
	marketCap := entry[2]
	period := entry[3]
	periodEnd := entry[4]
	guidRange := entry[5]
	midGuid := entry[6]
	cons := entry[7]
	percentToHighPoint := entry[8]

	return &GuidanceDataRow{
		Symbol:             symbol,
		Company:            company,
		MarketCap:          marketCap,
		Period:             period,
		PeriodEnd:          periodEnd,
		GuidRange:          guidRange,
		MidGuid:            midGuid,
		Cons:               cons,
		PercentToHighPoint: percentToHighPoint,
	}
}

func writeGuidanceData(w *os.File, data []*GuidanceDataRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(GuidanceDataRow), 4)
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
