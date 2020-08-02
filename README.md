**ip-api cache**

A proxy that caches query results from the ip-api.com pro endpoint.

- better performance via multiple keep-alive connections to ip-api servers
- finds the best ip-api PoP based on latency
- retries failed requests
- advanced caching for each response field
- automatically batches requests to reduce network requests

Only /json and /batch are supported.

After installing, change your applications to use ```http://127.0.0.1:8080``` instead of ```http(s)://pro.ip-api.com```

**Install on Debian**
- Get the latest version

```
go get -u github.com/ip-api/proxy
```

- Add user

```
adduser --system --disabled-password --disabled-login --home /opt/ip-api-proxy --group ip-api-proxy
```

- Copy the binary to ```/opt/ip-api-proxy/proxy```

- Create a config file

```
/opt/ip-api-proxy/config
```

Example configuration:

```
IP_API_KEY=your_api_key
LOG_OUTPUT=info
```

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

- Enable and start the service
```
systemctl enable ip-api-proxy
systemctl start ip-api-proxy
```

**Environment variables**

| Name             | Type     | Default                                         | Description |
| ---------------- | -------- | ----------------------------------------------- | ----------- |
| IP_API_KEY       | String   | *required*                                      | ip-api.com key |
| LISTEN           | String   | 127.0.0.1:8080                                  | ip:port to listen on |
| CACHE_TTL        | Duration | 24h                                             | For how long to cache entries |
| CACHE_SIZE       | Number   | 1073741824                                      | In memory cache size |
| RETRIES          | Number   | 4                                               | How many times to retry backend requests |
| POPS_REFRESH     | Duration | 1h                                              | How often to refresh the server locations  |
| BATCH_DELAY      | Duration | 10ms                                            | Max delay before sending a batch to the backend |
| LOG_OUTPUT       | String   | ""                                              | Set to "console" for console friendly output |
| LOG_LEVEL        | String   | ""                                              | Can be set to "info", "warn" or "error" to reduce log output |
| REVERSE_WORKERS  | Number   | 10                                              | How many workers to use for reverse lookups |
| REVERSE_PREFERGO | Bool     | true                                            | Prefer using Go's built-in DNS resolver |
