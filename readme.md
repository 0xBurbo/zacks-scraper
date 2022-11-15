
## Installation

Install go from https://go.dev/doc/install

Compile Repository
```bash
  go build
```

## Usage

Copy example.json to another file and change settings.
Each query in the queries section of config must have this shape:
```
{
    "id": "<query type>",
    "value": "either a number, or grade such as A, B, C, D",
    "operator": "either >=, <>, <=, =, same format as zacks.com"
}
```

Available query ids:
```
zacks_rank
zacks_industry_rank
value_score
growth_score
momentum_score
vgm_score
```


Run compiled binary with --config and --outDir flags
```bash
    ./zacks-scraper --config=/path/to/config --outDir=/path/to/output/folder
```

