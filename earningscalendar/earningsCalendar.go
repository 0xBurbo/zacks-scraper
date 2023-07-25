package earningscalendar

import (
	"encoding/json"
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
)

// Each tab in the calendar window
var EarningsCalendarTabs = map[string]int{
	"earnings":  1,
	"sales":     9,
	"guidance":  6,
	"revisions": 3,
	"dividends": 5,
	"splits":    4,
	// "transcripts": 8,
}

type earningsCalendarParams struct {
	start_date time.Time
	end_date   time.Time
	tabs       []string
}

// For unmarshaling raw response
type dataEntry []string
type earningsCalendarRawData struct {
	Data []dataEntry `json:"data"`
}

func RunEarningsCalendar(job *config.ScrapeJob, client *http.Client) error {

	params, err := parseJobParameters(job.Parameters)
	if err != nil {
		return err
	}

	currentDate := time.Now()
	job.OutDir = filepath.Join(job.OutDir, currentDate.Format("200601021504"))
	os.MkdirAll(job.OutDir, 0755)

	// Collect and write data
	temp := params.start_date
	for params.end_date.Sub(temp) >= 0 {
		for _, tab := range params.tabs {
			// Fetch data
			body, err := getEarningsCalendarData(temp, tab, client)
			if err != nil {
				return err
			}

			// Parse and save data
			data, err := parseEarningsCalendarBody(body)
			if err != nil {
				log.Printf("error parsing earnings data: %e", err)
			}

			// Create output file
			fileName := temp.Format("20060102150405") + "_" + tab + ".parquet"
			w, err := os.Create(filepath.Join(job.OutDir, fileName))
			if err != nil {
				return fmt.Errorf("failed to create local file: %e", err)
			}

			switch tab {
			case "earnings":
				rows := parseEarningsData(data)
				err = writeEarningsData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			case "sales":
				rows := parseSalesData(data)
				err = writeSalesData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			case "guidance":
				rows := parseGuidanceData(data)
				err = writeGuidanceData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			case "revisions":
				rows := parseRevisionsData(data)
				err = writeRevisionsData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			case "dividends":
				rows := parseDividendsData(data)
				err = writeDividendsData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			case "splits":
				rows := parseSplitsData(data)
				err = writeSplitsData(w, rows)
				if err != nil {
					log.Printf("error writing earnings data: %e", err)
				}
			}

			w.Close()
		}

		temp = temp.Add(24 * time.Hour)
	}

	return nil
}

// Parses arguments from config yaml
func parseJobParameters(parameters []map[string]interface{}) (*earningsCalendarParams, error) {
	var start_date time.Time
	var end_date time.Time
	var tabs []string

	for _, p := range parameters {
		if t, ok := p["start_date_offset"]; ok {
			offset, err := strconv.Atoi(t.(string))
			if err != nil {
				return nil, err
			}

			start_date = time.Now().AddDate(0, 0, offset)
		}

		if t, ok := p["end_date_offset"]; ok {
			offset, err := strconv.Atoi(t.(string))
			if err != nil {
				return nil, err
			}

			end_date = time.Now().AddDate(0, 0, offset)
		}

		if t, ok := p["start_date"]; ok {
			if t.(string) == "NOW" {
				start_date = time.Now().Add(1 * time.Hour)
				end_date = time.Now().Add(1 * time.Hour)
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

		if t, ok := p["tabs"]; ok {
			tInt := t.([]interface{})
			for _, v := range tInt {
				tabs = append(tabs, v.(string))
			}
		}
	}

	if len(tabs) == 0 {
		// NOTE: Excluded transcripts
		tabs = []string{"earnings", "sales", "guidance", "revisions", "dividends", "splits"}
	}

	return &earningsCalendarParams{
		start_date: start_date,
		end_date:   end_date,
		tabs:       tabs,
	}, nil
}

// Fetches raw earnings calendar data from Zacks
func getEarningsCalendarData(timestamp time.Time, tab string, client *http.Client) ([]byte, error) {
	u, err := url.Parse("https://www.zacks.com/includes/classes/z2_class_calendarfunctions_data.php")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("calltype", "eventscal")
	q.Set("date", strconv.Itoa(int(timestamp.Unix())))
	q.Set("type", strconv.Itoa(EarningsCalendarTabs[tab]))
	q.Set("search_trigger", "0")

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", "text/plain, */*; q=0.01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return bodyBytes, nil
	default:
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
}

// parses data returned from Earnings tab
func parseEarningsCalendarBody(body []byte) (*earningsCalendarRawData, error) {
	bodyString := string(body)

	// Remove javascript so body is in parsable JSON format
	parsable := strings.TrimPrefix(bodyString, `window.app_data = `)

	// Parse JSON
	rawData := &earningsCalendarRawData{}
	err := json.Unmarshal([]byte(parsable), rawData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %e", err)
	}

	return rawData, nil
}
