**Environment variables**

| Name    | Type   | Default | Description |
| ------- | ------ | ----------- |
| IP_API_KEY | String | *required* | ip-api.com key |
| LISTEN | String | 127.0.0.1:8080 | ip:port to listen on |
| CACHE_TTL | Duration | 24h | For how long to cache entries |
| CACHE_SIZE | Number | 1073741824 | In memory cache size |
| RETRIES | Number | 4 | How many times to retry backend requests |
| POPS_URL | Url | https://d2e7s0viy93a0y.cloudfront.net/pops.json | Endpoint to fetch server locations |
| POPS_REFRESH | Duration | 1h | How often to refresh the server locations |
