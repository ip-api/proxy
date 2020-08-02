**Install on Debian**

- Add user

```adduser --system --disabled-password --disabled-login --home /opt/ip-api-proxy --group ip-api-proxy```

- Create a systemd file
```[Unit]
Description=ip-api proxy server
Documentation=https://github.com/ip-api/proxy
After=network.target

[Service]
PermissionsStartOnly=true
LimitNOFILE=1048576
LimitNPROC=512
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE
NoNewPrivileges=true
User=ip-api-proxy
WorkingDirectory=/opt/ip-api-proxy
EnvironmentFile=/opt/ip-api-proxy/config
ExecStart=/opt/ip-api-proxy
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

- Create a config file

```/opt/ip-api-proxy/config```


**Environment variables**

| Name             | Type     | Default                                         | Description |
| ---------------- | -------- | ----------------------------------------------- | ----------- |
| IP_API_KEY       | String   | *required*                                      | ip-api.com key |
| LISTEN           | String   | 127.0.0.1:8080                                  | ip:port to listen on |
| CACHE_TTL        | Duration | 24h                                             | For how long to cache entries |
| CACHE_SIZE       | Number   | 1073741824                                      | In memory cache size |
| RETRIES          | Number   | 4                                               | How many times to retry backend requests |
| POPS_URL         | Url      | https://d2e7s0viy93a0y.cloudfront.net/pops.json | Endpoint to fetch server locations |
| POPS_REFRESH     | Duration | 1h                                              | How often to refresh the server locations  |
| BATCH_DELAY      | Duration | 10ms                                            | Max delay before sending a batch to the backend |
| LOG_OUTPUT       | String   | ""                                              | Set to "console" for console friendly output |
| LOG_LEVEL        | String   | ""                                              | Can be set to "info", "warn" or "error" to reduce log output |
| REVERSE_WORKERS  | Number   | 10                                              | How many workers to use for reverse lookups |
| REVERSE_PREFERGO | Bool     | true                                            | Prefer using Go's built-in DNS resolver |
