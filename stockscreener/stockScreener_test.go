package stockscreener

import (
	"fmt"
	"testing"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/util"
	"github.com/iamburbo/zacks-scraper/zacks"
)

func TestRunScreen(t *testing.T) {
	cfg, err := config.LoadConfigFile("../config.yml")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cfg)

	client := util.GetTestClient()

	err = zacks.LogIn(client, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var job *config.ScrapeJob = nil
	for _, j := range cfg.Jobs {
		if j.JobType == "stock_screener" {
			job = &j
			break
		}
	}

	err = RunStockScreener(job, client)
	if err != nil {
		t.Fatal(err)
	}
}
