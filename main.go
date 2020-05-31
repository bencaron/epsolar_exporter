package main

import (
	"fmt"
	"log"
	"net/http"

	collector "github.com/bencaron/epsolar_tracer_exporter"
	"github.com/jens18/gotracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	panelVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "epsolar_panel_voltage",
		Help: "Solar panel voltage (Volt).",
	})

	panelCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "epsolar_panel_current",
		Help: "Solar panel current (Amp).",
	})
	panelPower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "epsolar_panel_power",
		Help: "Solar panel power (Watts).",
	})

// 	// TracerStatus contain status information read from Tracer
// type TracerStatus struct {
// 	ArrayVoltage           float32   `json:"pvv"`     // Solar panel voltage, (V)
// 	ArrayCurrent           float32   `json:"pvc"`     // Solar panel current, (A)
// 	ArrayPower             float32   `json:"pvp"`     // Solar panel power, (W)
// 	BatteryVoltage         float32   `json:"bv"`      // Battery voltage, (V)
// 	BatteryCurrent         float32   `json:"bc"`      // Battery current, (A)
// 	BatterySOC             int32     `json:"bsoc"`    // Battery state of charge, (%)
// 	BatteryTemp            float32   `json:"btemp"`   // Battery temperatur, (C)
// 	BatteryMaxVoltage      float32   `json:"bmaxv"`   // Battery maximum voltage, (V)
// 	BatteryMinVoltage      float32   `json:"bminv"`   // Battery lowest voltage, (V)
// 	DeviceTemp             float32   `json:"devtemp"` // Tracer temperature, (C)
// 	LoadVoltage            float32   `json:"lv"`      // Load voltage, (V)
// 	LoadCurrent            float32   `json:"lc"`      // Load current, (A)
// 	LoadPower              float32   `json:"lp"`      // Load power, (W)
// 	Load                   bool      `json:"load"`    // Shows whether load is on or off
// 	EnergyConsumedDaily    float32   `json:"ecd"`     // Tracer calculated daily consumption, (kWh)
// 	EnergyConsumedMonthly  float32   `json:"ecm"`     // Tracer calculated monthly consumption, (kWh)
// 	EnergyConsumedAnnual   float32   `json:"eca"`     // Tracer calculated annual consumption, (kWh)
// 	EnergyConsumedTotal    float32   `json:"ect"`     // Tracer calculated total consumption, (kWh)
// 	EnergyGeneratedDaily   float32   `json:"egd"`     // Tracer calculated daily power generation, (kWh)
// 	EnergyGeneratedMonthly float32   `json:"egm"`     // Tracer calculated monthly power generation, (kWh)
// 	EnergyGeneratedAnnual  float32   `json:"ega"`     // Tracer calculated annual power generation, (kWh)
// 	EnergyGeneratedTotal   float32   `json:"egt"`     // Tracer calculated total power generation, (kWh)
// 	Timestamp              time.Time `json:"t"`
//}

)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(panelVoltage)
	prometheus.MustRegister(panelCurrent)
	prometheus.MustRegister(panelPower)
	prometheus.MustRegister(collector.NewMyCollector())
}

func main() {

	tracer, err := gotracer.Status("/dev/ttyUSB0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tracer)

	panelVoltage.Set(float64(tracer.ArrayVoltage))
	panelCurrent.Set(float64(tracer.ArrayCurrent))
	panelPower.Set(float64(tracer.ArrayPower))

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":6969", nil)

}
