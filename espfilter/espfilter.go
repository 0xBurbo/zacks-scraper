package espfilter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/util"
)

type EspFilterParameters struct {
	FilterType             string
	EspCheckboxes          []int
	ZacksRankCheckboxes    []int
	SurpCheckboxes         []int
	ReportingDateChecboxes []int
}

func RunEspFilter(job *config.ScrapeJob, client *http.Client) error {
	if job.JobType != "esp_filter" {
		return fmt.Errorf("invalid job type: %v", job)
	}

	// Parse parameters
	params := parseJobParameters(job.Parameters)

	body, err := filterRequest(client, params)
	if err != nil {
		return err
	}

	data, err := convertBodyToCSV(body)
	if err != nil {
		return err
	}

	// Write data to output directory
	fileName := time.Now().Format("20060102") + ".csv"
	f, err := os.Create(filepath.Join(job.OutDir, fileName))
	if err != nil {
		log.Fatalf("error creating output file: %v", err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range data {
		if err = w.Write(record); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	return nil
}

func parseJobParameters(parameters []map[string]interface{}) *EspFilterParameters {
	var filterType string
	espCheckboxes := []int{}
	zacksRankCheckboxes := []int{}
	surpCheckboxes := []int{}
	reportingDateChecboxes := []int{}
	for _, p := range parameters {
		if t, ok := p["filter_type"]; ok {
			filterType = t.(string)
		} else if values, ok := p["esp_checkboxes"]; ok {
			vArr := values.([]interface{})
			for _, v := range vArr {
				espCheckboxes = append(espCheckboxes, v.(int))
			}
		} else if values, ok := p["zacks_rank_checkboxes"]; ok {
			vArr := values.([]interface{})
			for _, v := range vArr {
				zacksRankCheckboxes = append(zacksRankCheckboxes, v.(int))
			}
		} else if values, ok := p["surp_checkboxes"]; ok {
			vArr := values.([]interface{})
			for _, v := range vArr {
				surpCheckboxes = append(surpCheckboxes, v.(int))
			}
		} else if values, ok := p["reporting_date_checkboxes"]; ok {
			vArr := values.([]interface{})
			for _, v := range vArr {
				reportingDateChecboxes = append(reportingDateChecboxes, v.(int))
			}
		}
	}

	return &EspFilterParameters{
		FilterType:             filterType,
		EspCheckboxes:          espCheckboxes,
		ZacksRankCheckboxes:    zacksRankCheckboxes,
		SurpCheckboxes:         surpCheckboxes,
		ReportingDateChecboxes: reportingDateChecboxes,
	}
}

func writeFilterQuery(parameters *EspFilterParameters) (string, error) {
	w := url.Values{}

	for _, i := range parameters.EspCheckboxes {
		w.Add("filter_checklist[]", "1#"+strconv.Itoa(i))
	}

	for _, i := range parameters.ZacksRankCheckboxes {
		w.Add("filter_checklist[]", "2#"+strconv.Itoa(i))
	}

	for _, i := range parameters.SurpCheckboxes {
		w.Add("filter_checklist[]", "3#"+strconv.Itoa(i))
	}

	for _, i := range parameters.ReportingDateChecboxes {
		w.Add("filter_checklist[]", "5#"+strconv.Itoa(i))
	}

	if parameters.FilterType == "buys" {
		w.Set("hd_esp_type", "1")
	} else if parameters.FilterType == "sells" {
		w.Set("hd_esp_type", "2")
	} else {
		return "", fmt.Errorf("esp filter: unknown filterType %v", parameters.FilterType)
	}

	return w.Encode(), nil
}

func filterRequest(client *http.Client, parameters *EspFilterParameters) ([]byte, error) {

	filterUrl, err := url.Parse("https://www.zacks.com/esp/esp_buysell_data_handler.php")
	if err != nil {
		return nil, err
	}

	// Construct filter queries
	buf := new(bytes.Buffer)
	body, err := writeFilterQuery(parameters)
	if err != nil {
		return nil, err
	}
	buf.Write([]byte(body))

	req, err := http.NewRequest("POST", filterUrl.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("content-type", `application/x-www-form-urlencoded; charset=UTF-8;`)

	cookieUrl, _ := url.Parse("https://www.zacks.com")
	for _, c := range client.Jar.Cookies(cookieUrl) {
		req.AddCookie(c)
	}
	req.AddCookie(&http.Cookie{
		Name:  "CURRENT_POST",
		Value: "edit_criteria",
	})

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		return bodyBytes, nil
	default:
		return nil, errors.New("status code: " + strconv.Itoa(resp.StatusCode))
	}
}

func convertBodyToCSV(body []byte) ([][]string, error) {
	// Header
	csvArray := [][]string{}
	csvArray = append(csvArray, []string{"Symbol", "Company", "ESP", "Most Accurate Estimate", "Consensus Estimate", "Price", "Zacks Rank", "% Surprise (Last Qtr.)", "Reporting Date"})

	type espFilterResponse struct {
		Data [][]string `json:"data"`
	}

	data := &espFilterResponse{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	for _, d := range data.Data {

		symbol := util.FindStringInBetween(`<span class="hoverquote-symbol">`, `<span class="sr-only">`, d[0])
		companyName := util.FindStringInBetween(`>`, `</a>`, d[1])
		esp := util.FindStringInBetween(`>`, `</span>`, d[2])
		zacksRank := util.FindStringInBetween(`>`, `</span>`, d[6])
		percentSurprise := util.FindStringInBetween(`>`, `</span>`, d[7])

		csvArray = append(csvArray, []string{symbol, companyName, esp, d[3], d[4], d[5], zacksRank, percentSurprise, d[8]})
	}

	return csvArray, nil
}
