
## Installation

Install go from https://go.dev/doc/install

Compile Repository
```bash
  go build
```

## Usage

Available stock screener query ids:
```
zacks_rank
zacks_industry_rank
value_score
growth_score
momentum_score
vgm_score
earnings_esp
52_week_high
market_cap
last_eps_surprise
p_n_e
num_brokers
optionable
percent_change_f1
div_yield
avg_volume
```


Run compiled binary with --config flag
```bash
    ./zacks-scraper --config=/path/to/config
```

