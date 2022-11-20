package main

import (
	"log"
	"net/http"
	"os"
	"switchbot-exporter/pkg/switchbot"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	token := os.Getenv("SWITCHBOT_TOKEN")

	if token == "" {
		log.Println("The environment variable SWITCHBOT_TOKEN is empty.")
		os.Exit(1)
	}

	client := switchbot.NewClient(token)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {

		deviceID := r.FormValue("target")

		if deviceID == "" {
			http.Error(w, "Target parameter is missing", http.StatusBadRequest)
			return
		}

		deviceName, err := client.GetDeviceName(deviceID)

		if err != nil {
			log.Println(err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		humidity := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "switchbot",
			Name:      "humidity",
			Help:      "Humidity measured with a Switchbot thermo-hygrometer",
		}, []string{"device_name", "device_id"})
		temperature := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "switchbot",
			Name:      "temperature",
			Help:      "Temperature measured with a Switchbot thermo-hygrometer",
		}, []string{"device_name", "device_id"})

		reg := prometheus.NewRegistry()
		reg.MustRegister(humidity, temperature)

		temperatureValue, humidityValue, err := client.GetThermometerValue(deviceID)

		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		humidity.WithLabelValues(deviceName, deviceID).Set(float64(humidityValue))
		temperature.WithLabelValues(deviceName, deviceID).Set(float64(temperatureValue))

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
