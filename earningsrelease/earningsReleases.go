package earningsrelease

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type EarningsReleaseParams struct {
	start_date time.Time
	end_date   time.Time
}

type RawEarningsReleaseRow struct {
	Symbol             string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Company            string `parquet:"name=company, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	ReportTime         string `parquet:"name=reportTime, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Estimate           string `parquet:"name=estimate, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Reported           string `parquet:"name=reported, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Surprise           string `parquet:"name=surprise, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	CurrentPrice       string `parquet:"name=currentPrice, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	PricePercentChange string `parquet:"name=pricePercentChange, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func RunEarningsRelease(job *config.ScrapeJob, client *http.Client) error {
	params, err := parseJobParameters(job.Parameters)
	if err != nil {
		return fmt.Errorf("error parsing parameters: %e", err)
	}

	temp := params.start_date
	for params.end_date.Sub(temp) >= 0 {
		// Fetch data
		body, err := getEarningsRelease(temp, client)
		if err != nil {
			return err
		}

		// Parse data
		parsedRows := parseEarningReleaseBody(body, temp)

		// Write to parquet
		fileName := temp.Format("20060102") + ".parquet"
		w, err := os.Create(filepath.Join(job.OutDir, fileName))
		if err != nil {
			return fmt.Errorf("failed to create local file: %e", err)
		}
		writeToParquet(w, parsedRows)
		w.Close()

		// Move on to next day
		temp = temp.Add(24 * time.Hour)
	}

	return nil
}

func parseJobParameters(parameters []map[string]interface{}) (*EarningsReleaseParams, error) {
	var start_date time.Time
	var end_date time.Time

	for _, p := range parameters {
		if t, ok := p["start_date"]; ok {
			if t.(string) == "NOW" {
				start_date = time.Now()
				end_date = time.Now()
			} else {
				parsed, err := time.Parse("2006-01-02", t.(string))
				if err != nil {
					return nil, err
				}
				parsed = parsed.Add(6 * time.Hour)

				start_date = parsed
			}
		}

		if t, ok := p["end_date"]; ok {
			if t.(string) == "NOW" {
				start_date = time.Now()
				end_date = time.Now()
			} else {
				parsed, err := time.Parse("2006-01-02", t.(string))
				if err != nil {
					return nil, err
				}
				parsed = parsed.Add(6 * time.Hour)

				end_date = parsed
			}
		}
	}

	return &EarningsReleaseParams{
		start_date: start_date,
		end_date:   end_date,
	}, nil
}

func getEarningsRelease(timestamp time.Time, client *http.Client) ([]byte, error) {
	u, err := url.Parse("https://www.zacks.com/research/earnings/earning_export.php")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("timestamp", strconv.Itoa(int(timestamp.Unix())))
	q.Set("tab_id", "1")

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	default:
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

}

func parseEarningReleaseBody(body []byte, timestamp time.Time) []*RawEarningsReleaseRow {

	parsedRows := []*RawEarningsReleaseRow{}

	rows := strings.Split(string(body), "\n")
	for _, row := range rows[1:] {
		items := strings.Split(row, "\t")

		if len(items) == 9 {
			parsedItem := &RawEarningsReleaseRow{
				Symbol:             items[0],
				Company:            items[1],
				ReportTime:         items[2],
				Estimate:           items[3],
				Reported:           items[4],
				Surprise:           items[5],
				CurrentPrice:       items[6],
				PricePercentChange: items[7],
			}

			parsedRows = append(parsedRows, parsedItem)

		}

	}

	return parsedRows
}

func writeParsedRowsToCSV(rows []*RawEarningsReleaseRow, outDir string, timestamp time.Time) {
	fileName := timestamp.Format("20060102") + ".csv"
	f, err := os.Create(filepath.Join(outDir, fileName))
	if err != nil {
		log.Fatalf("error creating output file: %v", err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()

	err = w.Write([]string{"Symbol", "Company", "Report Time", "Estimate", "Reported", "Surprise", "Current Price", "Price % Change"})
	if err != nil {
		log.Fatalln("error writing header record", err)
	}

	for _, record := range rows {
		if err = w.Write([]string{record.Symbol, record.Company, record.ReportTime, record.Estimate, record.Reported, record.Surprise, record.CurrentPrice, record.PricePercentChange}); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
}

func writeToParquet(w *os.File, data []*RawEarningsReleaseRow) error {
	pw, err := writer.NewParquetWriterFromWriter(w, new(RawEarningsReleaseRow), 4)
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
