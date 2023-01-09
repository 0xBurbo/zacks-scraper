package stockscreener

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/util"
)

func RunStockScreener(job *config.ScrapeJob, client *http.Client) error {
	if job.JobType != "stock_screener" {
		return fmt.Errorf("invalid job type: %v", job)
	}

	// Send requests
	prefix := "an error occured while"
	parsedStockScreenerPage, err := getStockScreenerPage(client)
	if err != nil {
		return fmt.Errorf("%v fetching stock screener page: %w", prefix, err)
	}

	err = getScreenerFromApi(client, parsedStockScreenerPage)
	if err != nil {
		return fmt.Errorf("%v fetching screener API page: %w", prefix, err)
	}

	err = resetStockScreenerParam(client)
	if err != nil {
		return fmt.Errorf("%v resetting query params: %w", prefix, err)
	}

	err = queryScreenerApi(client, job.Parameters)
	if err != nil {
		return fmt.Errorf("%v sending query: %w", prefix, err)
	}

	data, err := downloadData(client, parsedStockScreenerPage)
	if err != nil {
		return fmt.Errorf("%v resetting query params: %w", prefix, err)
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

type parsedStockScreenerHomePage struct {
	CKey string
}

func parseStockScreenerHomePage(body string) (*parsedStockScreenerHomePage, error) {
	substr := util.FindStringInBetween(`<iframe style="" title="Stock Screener " id="screenerContent" src="`, `" scrolling="yes" allowfullscreen></iframe>`, body)
	parsed, err := url.Parse(substr)
	if err != nil {
		return nil, err
	}

	ckey := parsed.Query().Get("c_key")
	if ckey == "" {
		return nil, errors.New("no c_key found")
	}
	return &parsedStockScreenerHomePage{
		CKey: ckey,
	}, nil
}

// fetch screener page from frontend to set important cookies
func getStockScreenerPage(client *http.Client) (*parsedStockScreenerHomePage, error) {
	screenerUrl, err := url.Parse("https://www.zacks.com/screening/stock-screener?icid=home-home-nav_tracking-zcom-main_menu_wrapper-stock_screener")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", screenerUrl.String(), nil)
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	parsed, err := parseStockScreenerHomePage(string(bodyBytes))
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		return parsed, nil
	default:
		return nil, errors.New("status code: " + strconv.Itoa(resp.StatusCode))
	}
}

// Retrieves screener prompt homepage from backend API. Necessary to authorize future requests
func getScreenerFromApi(client *http.Client, parsed *parsedStockScreenerHomePage) error {
	screenerUrl, err := url.Parse(`https://screener-api.zacks.com/`)
	if err != nil {
		return err
	}

	q := screenerUrl.Query()
	q.Set("scr_type", "stock")
	q.Set("c_id", "zacks")
	q.Set("c_key", parsed.CKey)
	q.Set("ref", "screening")

	screenerUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", screenerUrl.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	cookieUrl, _ := url.Parse("https://www.zacks.com")
	for _, c := range client.Jar.Cookies(cookieUrl) {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return errors.New("status code " + strconv.Itoa(resp.StatusCode))
	}
}

// Sends query to stock screener api via multipart form data
func queryScreenerApi(client *http.Client, parameters []map[string]interface{}) error {
	screenApiUrl, err := url.Parse("https://screener-api.zacks.com/getrunscreendata.php")
	if err != nil {
		return err
	}

	// Query body writer
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	boundary := `----WebKitFormBoundarynYAVaZAwVXgNHDGd`
	writer.SetBoundary(boundary)

	// Queries
	err = WriteQuery(writer, parameters)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", screenApiUrl.String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("sec-ch-ua", `"Chromium";v="106", "Google Chrome";v="106", "Not;A=Brand";v="99"`)
	req.Header.Set("accept", "*/*")
	req.Header.Set("content-type", `multipart/form-data; boundary=`+boundary)
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("origin", "https://screener-api.zacks.com")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("referer", "https://screener-api.zacks.com/?scr_type=stock&c_id=zacks&c_key=0675466c5b74cfac34f6be7dc37d4fe6a008e212e2ef73bdcd7e9f1f9a9bd377&ecv=4MTNzETOygTM&ref=screening")

	// Cookies won't auto add since login was from a different host
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
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return errors.New("Status code: " + strconv.Itoa(resp.StatusCode))
	}
}

// Called by browser - reset params just in case
func resetStockScreenerParam(client *http.Client) error {
	resetParamUrl, err := url.Parse(`https://screener-api.zacks.com/reset_param.php`)
	if err != nil {
		return err
	}

	currTimestamp := time.Now()

	q := resetParamUrl.Query()
	q.Add("_", strconv.FormatInt(currTimestamp.UnixMilli(), 10))
	q.Add("mode", "new")

	resetParamUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", resetParamUrl.String(), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return errors.New("status code " + strconv.Itoa(resp.StatusCode))
	}

}

// Downloads query in CSV format
func downloadData(client *http.Client, parsed *parsedStockScreenerHomePage) ([][]string, error) {
	downloadUrl, err := url.Parse(`https://screener-api.zacks.com/export.php`)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", downloadUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("sec-ch-ua", `"Chromium";v="106", "Google Chrome";v="106", "Not;A=Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("sec-fetch-dest", "iframe")
	req.Header.Set("referer", "https://screener-api.zacks.com/?scr_type=stock&c_id=zacks&c_key="+parsed.CKey+"&ecv=4MTNzETOygTM&ref=screening")
	req.Header.Set("accept-language", "en-US,en;q=0.9")

	// Cookies won't auto add since login was from a different host
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

	switch resp.StatusCode {
	case 200:
		reader := csv.NewReader(resp.Body)
		data, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, errors.New("Status code " + strconv.Itoa(resp.StatusCode))
	}
}
