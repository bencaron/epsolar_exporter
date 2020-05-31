package main

import (
	"log"
	"sync"

	"github.com/jens18/gotracer"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "solar"
)

var (
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

type solarCollector struct {
	mutex sync.Mutex

	scrapeFailures prometheus.Counter

	panelVoltage *prometheus.Desc
	panelCurrent *prometheus.Desc
	panelPower   *prometheus.Desc

	batteryVoltage    *prometheus.Desc
	batteryCurrent    *prometheus.Desc
	batterySOC        *prometheus.Desc
	batteryTemp       *prometheus.Desc
	batteryMaxVoltage *prometheus.Desc
	batteryMinVoltage *prometheus.Desc

	DeviceTemp *prometheus.Desc

	loadActive  *prometheus.Desc
	loadVoltage *prometheus.Desc
	loadCurrent *prometheus.Desc
	loadPower   *prometheus.Desc

	energyConsumedDaily    *prometheus.Desc
	energyConsumedMonthly  *prometheus.Desc
	energyConsumedAnnual   *prometheus.Desc
	energyConsumedTotal    *prometheus.Desc
	energyGeneratedDaily   *prometheus.Desc
	energyGeneratedMonthly *prometheus.Desc
	energyGeneratedAnnual  *prometheus.Desc
	energyGeneratedTotal   *prometheus.Desc
}

func newSolarCollector() *solarCollector {
	return &solarCollector{
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "controller_comm_failures_total",
			Help:      "Number of communications errors while connecting to the solar controller.",
		}),
		panelVoltage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "panel_voltage"),
			"Solar panel voltage (V).",
			nil, // no labels yet
			nil,
		),
		panelCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "panel_current"),
			"Solar panel current (A).",
			nil, // no labels yet
			nil,
		),
		panelPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "panel_power"),
			"Solar panel power (W).",
			nil, // no labels yet
			nil,
		),
		batteryVoltage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_voltage"),
			"Battery voltage (V).",
			nil, // no labels yet
			nil,
		),
		batteryCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_current"),
			"Battery current (A).",
			nil, // no labels yet
			nil,
		),
		batterySOC: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_soc"),
			"Battery State of Charge (%).",
			nil, // no labels yet
			nil,
		),
		batteryTemp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_temp"),
			"Battery temperature (controller sensor) (Celcius).",
			nil, // no labels yet
			nil,
		),
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *solarCollector) Describe(ch chan<- *prometheus.Desc) {

	// Describe the Collector's member that are of type Desc
	ds := []*prometheus.Desc{
		c.panelVoltage,
	}

	for _, d := range ds {
		ch <- d
	}
	// Describe the other types
	c.scrapeFailures.Describe(ch)
}

// Collect gather the metrics values and sends them.
// The call is protected from concurrent collects with a mutex lock.
func (c *solarCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()
	if err := c.collect(ch); err != nil {
		log.Printf("Error getting solar controller data: %s", err)
		c.scrapeFailures.Inc()
		c.scrapeFailures.Collect(ch)
	}
	return
}

// collect will execute the actual data collection
func (c *solarCollector) collect(ch chan<- prometheus.Metric) error {
	// fetch the status of the controller
	tracer, err := gotracer.Status("/dev/ttyUSB0")
	if err != nil {
		return err
	}
	/*
		// fmt.Println(tracer)
		panelVoltage.Set(float64(tracer.ArrayVoltage))
		panelCurrent.Set(float64(tracer.ArrayCurrent))
		panelPower.Set(float64(tracer.ArrayPower))
	*/
	ch <- prometheus.MustNewConstMetric(
		c.panelVoltage,
		prometheus.CounterValue,
		float64(tracer.ArrayVoltage),
	)
	ch <- prometheus.MustNewConstMetric(
		c.panelCurrent,
		prometheus.CounterValue,
		float64(tracer.ArrayCurrent),
	)
	// Bloc pou les batteries
	ch <- prometheus.MustNewConstMetric(
		c.batteryCurrent,
		prometheus.CounterValue,
		float64(tracer.BatteryCurrent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.batterySOC,
		prometheus.CounterValue,
		float64(tracer.BatterySOC),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryTemp,
		prometheus.CounterValue,
		float64(tracer.BatteryTemp),
	)
	return nil
}

// To hook in the collector: prometheus.MustRegister(NewsolarCollector())
