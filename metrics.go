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
			[]string{"plant_id", "last_update"}, nil),

		plantPower: prometheus.NewDesc(
			"plant_power",
			"Current power produced by the monitored plant",
			[]string{"plant_id"}, nil),
	}
}

func (m metrics) Collect(ch chan<- prometheus.Metric) {
	login(*username, *password)
	plantIds := getPlantIds()

	for _, plantId := range plantIds {
		plant := getPlantData(plantId)

		lastUpdate := plant["last_data_time"].(string)
		plantPower, _ := strconv.ParseFloat(plant["real_power"].(string), 64)

		ch <- prometheus.MustNewConstMetric(m.plantInfo, prometheus.GaugeValue, 1, fmt.Sprintf("%d", plantId), lastUpdate)
		ch <- prometheus.MustNewConstMetric(m.plantPower, prometheus.GaugeValue, plantPower, fmt.Sprintf("%d", plantId))
	}
}

func (m metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.plantInfo
}
