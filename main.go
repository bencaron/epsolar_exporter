package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// Metrics have to be registered to be exposed:
	/*	prometheus.MustRegister(panelVoltage)
		prometheus.MustRegister(panelCurrent)
		prometheus.MustRegister(panelPower)
	*/
	prometheus.MustRegister(newSolarCollector())
}

type config struct {
	listen string
	port   string
}

func newConfig() config {
	c := config{}
	c.port = "6969"
	c.listen = "0.0.0.0"
	return c
}

func main() {
	var conf = newConfig()

	log.Println("Starting epsolar_exporter")
	log.Println("Listening on 0.0.0.0:6969")
	log.Println(fmt.Sprintf("Listening on %s:%s", conf.listen, conf.port))
	http.Handle("/metrics", promhttp.Handler())

	//http.ListenAndServe(fmt.Sprintf("%s:%s", conf.listen, conf.port), nil)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", conf.listen, conf.port), nil))
}
