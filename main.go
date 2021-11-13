package main

import (
	"context"
	"log"
	"net/http"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config stores the parameters used to fetch the data
type Config struct {
	PollingInterval time.Duration `env:"OWM_POLLING_INTERVAL" envDefault:"60s"`
	RequestTimeout  time.Duration `env:"OWM_TIMEOUT" envDefault:"1s"`
	APIKey          string        `env:"OWM_API_KEY"` // APIKey delivered by Openweathermap
	Location        string        `env:"OWM_LOCATION" envDefault:"Lille,FR"`
	Language        string        `env:"OWM_LANGUAGE" envDefault:"en"`
	ServerPort      string        `env:"SERVER_PORT" envDefault:"2112"`
	Units           string        `env:"OWM_UNITS" envDefault:"C"`
}

func loadMetrics(ctx context.Context, location string) <-chan error {
	errC := make(chan error)
	go func() {
		c := time.Tick(cfg.PollingInterval)
		for {
			select {
			case <-ctx.Done():
				return // returning not to leak the goroutine
			case <-c:
				client := &http.Client{
					Timeout: cfg.RequestTimeout,
				}

				w, err := owm.NewCurrent(cfg.Units, cfg.Language, cfg.APIKey, owm.WithHttpClient(client)) // (internal - OpenWeatherMap reference for kelvin) with English output
				if err != nil {
					errC <- err
					continue
				}

				err = w.CurrentByName(location)
				if err != nil {
					errC <- err
					continue
				}

				temp.WithLabelValues(location).Set(w.Main.Temp)

				pressure.WithLabelValues(location).Set(w.Main.Pressure)

				humidity.WithLabelValues(location).Set(float64(w.Main.Humidity))

				wind.WithLabelValues(location).Set(w.Wind.Speed)

				clouds.WithLabelValues(location).Set(float64(w.Clouds.All))

				rain.WithLabelValues(location).Set(w.Rain.OneH)

				var scraped_weather = w.Weather[0].Description
				if scraped_weather == last_weather {
					weather.WithLabelValues(location, scraped_weather).Set(1)
				} else {
					weather.WithLabelValues(location, scraped_weather).Set(1)
					weather.WithLabelValues(location, last_weather).Set(0)
					last_weather = scraped_weather
				}
				log.Println(w.Weather[0].Description)
				log.Println("scraping OK for ", location)
			}
		}
	}()
	return errC
}

var (
	cfg = Config{}

	temp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "temperature_celsius",
		Help:      "Temperature in °C",
	}, []string{"location"})

	pressure = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "pressure_hpa",
		Help:      "Atmospheric pressure in hPa",
	}, []string{"location"})

	humidity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "humidity_percent",
		Help:      "Humidity in Percent",
	}, []string{"location"})

	wind = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "wind_mps",
		Help:      "Wind speed in m/s",
	}, []string{"location"})

	clouds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "cloudiness_percent",
		Help:      "Cloudiness in Percent",
	}, []string{"location"})

	rain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "rain",
		Help:      "Rain contents 1h",
	}, []string{"location"})

	weather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "weather",
		Help:      "The weather label.",
	}, []string{"location", "weather"})

	last_weather = ""
)

func main() {
	env.Parse(&cfg)

	prometheus.Register(temp)
	prometheus.Register(pressure)
	prometheus.Register(humidity)
	prometheus.Register(wind)
	prometheus.Register(clouds)
	prometheus.Register(rain)
	prometheus.Register(weather)

	errC := loadMetrics(context.TODO(), cfg.Location)
	go func() {
		for err := range errC {
			log.Println(err)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+cfg.ServerPort, nil)
}
