package zacks

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/iamburbo/zacks-scraper/config"
)

// Sends login request to set session cookie in cookie jar
func LogIn(client *http.Client, config *config.Config) error {
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
