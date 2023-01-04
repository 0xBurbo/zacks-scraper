package espfilter

import (
	"testing"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/util"
	"github.com/iamburbo/zacks-scraper/zacks"
)

func TestRunEspFilter(t *testing.T) {
	cfg, err := config.LoadConfigFile("../config.yml")
	if err != nil {
		t.Fatal(err)
	}

	client := util.GetTestClient()

	err = zacks.LogIn(client, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var job *config.ScrapeJob = nil
	for _, j := range cfg.Jobs {
		if j.JobType == "esp_filter" {
			job = &j
			break
		}
	}

	err = RunEspFilter(job, client)
	if err != nil {
		t.Fatal(err)
	}
}
