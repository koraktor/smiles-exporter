package main

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	plantInfo  *prometheus.Desc
	plantPower *prometheus.Desc
}

func newMetrics() metrics {
	return metrics{
		plantInfo: prometheus.NewDesc(
			"plant_info",
			"Basic information about the monitored plant",
			[]string{"plant_id", "name", "last_update", "max_power"}, nil),

		plantPower: prometheus.NewDesc(
			"plant_power",
			"Current power produced by the monitored plant",
			[]string{"plant_id"}, nil),
	}
}

func (m metrics) Collect(ch chan<- prometheus.Metric) {
	login(*username, *password)
	plantInfo := getPlants()

	for _, plant := range plantInfo {
		plantData := getPlantData(plant.Id)

		maxPower, _ := strconv.ParseFloat(plantData.Data.MaxPower, 64)
		maxPower *= 1000
		lastUpdate := plantData.Data.LastDataTime
		plantPower, _ := strconv.ParseFloat(plantData.Data.RealPower, 64)

		ch <- prometheus.MustNewConstMetric(m.plantInfo, prometheus.GaugeValue, 1, fmt.Sprintf("%.0f", plant.Id), plant.Name, lastUpdate, fmt.Sprintf("%.0f", maxPower))
		ch <- prometheus.MustNewConstMetric(m.plantPower, prometheus.GaugeValue, plantPower, fmt.Sprintf("%.0f", plant.Id))
	}
}

func (m metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.plantInfo
}
