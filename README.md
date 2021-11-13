# Openweather exporter for prometheus


Exporter for openweather API 


## Quickstart

Create an API key from https://openweathermap.org/.

Install dependancies with `go mod init` and `go get` and then build the binary.

```
go mod init github.com/blackrez/openweathermap_exporter
go get -d -v
go build
OWM_LOCATION=LONDON,UK  OWM_API_KEY=apikey ./openweathermap_exporter
```

Then add the scraper in prometheus

```
scrape_configs:
  - job_name: 'weather'

    # Scrape is configured for free usage.
    scrape_interval: 60s

    # Port is not yet configurable
    static_configs:
      - targets: ['localhost:2112']
```



## With Docker

The image is a multistage image, just launch as usual :

```
docker-compose up -d
```

## Configuration

```
OWM_API_KEY Your OWM API Key.
OWM_LOCATION Location you want to monitor. default: Lille,FR
OWM_UNITS Units: C (Celsius), F (Fahrenheit) or K (Kelvin). default: C
OWM_LANGUAGE Language to display in. default: en
OWM_POLLING_INTERVAL Interval at which OWM is polled. default: 60s
OWM_TIMEOUT Timeout for requests. default: 1s
SERVER_PORT Port where /metrics are published, and must match docker-compose port value. default: 2112
```