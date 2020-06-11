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

	deviceTemp *prometheus.Desc

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

// newSolarCollector makes a new solarCollector objects with fully initialized metrics descriptors
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
			"Battery temperature (external sensor) (Celcius).",
			nil, // no labels yet
			nil,
		),
		batteryMaxVoltage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_max_voltage"),
			"Maximum battery voltage (V).",
			nil, // no labels yet
			nil,
		),
		batteryMinVoltage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_min_voltage"),
			"Minimum battery voltage (V).",
			nil, // no labels yet
			nil,
		),

		deviceTemp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "device_temp"),
			"Device temperature (controller sensor) (Celcius).",
			nil, // no labels yet
			nil,
		),

		loadActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "load_active"),
			"Load output is active (bool)",
			nil, // no labels yet
			nil,
		),
		loadVoltage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "load_voltage"),
			"Load voltage (V).",
			nil, // no labels yet
			nil,
		),
		loadCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "load_current"),
			"Load current (A).",
			nil, // no labels yet
			nil,
		),
		loadPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "load_power"),
			"Load power (W).",
			nil, // no labels yet
			nil,
		),

		energyConsumedDaily: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_consumed_daily"),
			"Controller calculated daily consumption, (kWh)",
			nil, // no labels yet
			nil,
		),

		energyConsumedMonthly: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_consumed_monthly"),
			"Controller calculated monthly consumption, (kWh)",
			nil, // no labels yet
			nil,
		),
		energyConsumedAnnual: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_consumed_annual"),
			"Controller calculated annual consumption, (kWh)",
			nil, // no labels yet
			nil,
		),
		energyConsumedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_consumed_taotal"),
			"Controller calculated total consumption, (kWh)",
			nil, // no labels yet
			nil,
		),

		energyGeneratedDaily: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_generated_daily"),
			"Controller calculated daily power generation, (kWh)",
			nil, // no labels yet
			nil,
		),
		energyGeneratedMonthly: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_generated_monthly"),
			"Controller calculated monthly power generation, (kWh)",
			nil, // no labels yet
			nil,
		),
		energyGeneratedAnnual: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_generated_annual"),
			"Controller calculated annual power generation, (kWh)",
			nil, // no labels yet
			nil,
		),
		energyGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_generated_total"),
			"Controller calculated total power generation, (kWh)",
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
	 *  report the collected data
	 */

	// store boolean values as a float (1 == true, 0 == false)
	var loadIsActive float64
	// Panel array
	ch <- prometheus.MustNewConstMetric(
		c.panelVoltage,
		prometheus.GaugeValue,
		float64(tracer.ArrayVoltage),
	)
	ch <- prometheus.MustNewConstMetric(
		c.panelCurrent,
		prometheus.GaugeValue,
		float64(tracer.ArrayCurrent),
	)
	ch <- prometheus.MustNewConstMetric(
		c.panelPower,
		prometheus.GaugeValue,
		float64(tracer.ArrayPower),
	)

	// Batteries
	ch <- prometheus.MustNewConstMetric(
		c.batteryCurrent,
		prometheus.GaugeValue,
		float64(tracer.BatteryCurrent),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryVoltage,
		prometheus.GaugeValue,
		float64(tracer.BatteryVoltage),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batterySOC,
		prometheus.GaugeValue,
		float64(tracer.BatterySOC),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryTemp,
		prometheus.GaugeValue,
		float64(tracer.BatteryTemp),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryMinVoltage,
		prometheus.GaugeValue,
		float64(tracer.BatteryMinVoltage),
	)
	ch <- prometheus.MustNewConstMetric(
		c.batteryMaxVoltage,
		prometheus.GaugeValue,
		float64(tracer.BatteryMaxVoltage),
	)

	// Load output
	if tracer.Load {
		loadIsActive = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.loadActive,
		prometheus.GaugeValue,
		loadIsActive,
	)
	ch <- prometheus.MustNewConstMetric(
		c.loadVoltage,
		prometheus.GaugeValue,
		float64(tracer.LoadVoltage),
	)
	ch <- prometheus.MustNewConstMetric(
		c.loadCurrent,
		prometheus.GaugeValue,
		float64(tracer.LoadCurrent),
	)
	ch <- prometheus.MustNewConstMetric(
		c.loadPower,
		prometheus.GaugeValue,
		float64(tracer.LoadPower),
	)

	// controller infos
	ch <- prometheus.MustNewConstMetric(
		c.deviceTemp,
		prometheus.GaugeValue,
		float64(tracer.DeviceTemp),
	)

	// energy consumed
	ch <- prometheus.MustNewConstMetric(
		c.energyConsumedDaily,
		prometheus.GaugeValue,
		float64(tracer.EnergyConsumedDaily),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyConsumedMonthly,
		prometheus.GaugeValue,
		float64(tracer.EnergyConsumedMonthly),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyConsumedAnnual,
		prometheus.GaugeValue,
		float64(tracer.EnergyConsumedAnnual),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyConsumedTotal,
		prometheus.GaugeValue,
		float64(tracer.EnergyConsumedTotal),
	)
	// energy generated
	ch <- prometheus.MustNewConstMetric(
		c.energyGeneratedDaily,
		prometheus.GaugeValue,
		float64(tracer.EnergyGeneratedDaily),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyGeneratedMonthly,
		prometheus.GaugeValue,
		float64(tracer.EnergyGeneratedMonthly),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyGeneratedAnnual,
		prometheus.GaugeValue,
		float64(tracer.EnergyGeneratedAnnual),
	)
	ch <- prometheus.MustNewConstMetric(
		c.energyGeneratedTotal,
		prometheus.GaugeValue,
		float64(tracer.EnergyGeneratedTotal),
	)

	return nil
}
