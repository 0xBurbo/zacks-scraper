package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Sends login request to set session cookie in cookie jar
func LogIn(client *http.Client, config *zacksScraperInput) error {
	loginUrl, err := url.Parse("https://www.zacks.com")
	if err != nil {
		return err
	}

	q := loginUrl.Query()

	// TODO: Move credentials out of repo
	q.Add("force_login", "true")
	q.Add("username", config.Username)
	q.Add("password", config.Password)
	q.Add("remember_me", "off")

	loginUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", loginUrl.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36`)
	req.Header.Set("accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return errors.New("Login status: " + strconv.Itoa(resp.StatusCode))
	}
}

// fetch screener page from frontend to set important cookies
func GetStockScreenerPage(client *http.Client) (*parsedStockScreenerHomePage, error) {
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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
func GetScreenerFromApi(client *http.Client, parsed *parsedStockScreenerHomePage) error {
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
func QueryScreenerApi(client *http.Client, config *zacksScraperInput) error {
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
	WriteQuery(writer, config.Queries)

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
func ResetParam(client *http.Client) error {
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
func DownloadData(client *http.Client, parsed *parsedStockScreenerHomePage) ([][]string, error) {
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
