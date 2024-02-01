package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var timeZoneRegex = regexp.MustCompile("UTC([+-])(\\d{2})")

type metrics struct {
	plantEnergyTotal *prometheus.Desc
	plantInfo        *prometheus.Desc
	plantPower       *prometheus.Desc
}

func newMetrics() metrics {
	return metrics{
		plantEnergyTotal: prometheus.NewDesc(
			"smiles_plant_energy_total",
			"Total energy produced by a monitored plant",
			[]string{"plant_id"}, nil),

		plantInfo: prometheus.NewDesc(
			"smiles_plant_info",
			"Basic information about the monitored plant",
			[]string{"plant_id", "name", "last_update", "max_power"}, nil),

		plantPower: prometheus.NewDesc(
			"smiles_plant_power",
			"Current power produced by the monitored plant",
			[]string{"plant_id"}, nil),
	}
}

func (m metrics) Collect(ch chan<- prometheus.Metric) {
	login(*username, *password)
	plantInfo := getPlants()

	for _, plant := range plantInfo {
		plantData := getPlantData(plant.Id)

		energyToday, _ := strconv.ParseFloat(plantData.Data.EnergyToday, 64)
		energyTotal, _ := strconv.ParseFloat(plantData.Data.EnergyTotal, 64)
		maxPower, _ := strconv.ParseFloat(plantData.Data.MaxPower, 64)
		maxPower *= 1000

		timeZone := timeZoneRegex.FindAllStringSubmatch(plant.TimeZone, 1)[0]
		timeZoneNegative := timeZone[1] == "-"
		timeZoneOffset, _ := strconv.ParseInt(timeZone[2], 10, 64)
		if timeZoneNegative {
			timeZoneOffset *= -1
		}

		location := time.FixedZone(plant.TimeZone, int(timeZoneOffset)*3600)
		lastUpdate, _ := time.ParseInLocation(time.DateTime, plantData.Data.LastDataTime, location)
		plantPower, _ := strconv.ParseFloat(plantData.Data.RealPower, 64)

		ch <- prometheus.MustNewConstMetric(m.plantEnergyTotal, prometheus.CounterValue, energyTotal+energyToday, fmt.Sprintf("%.0f", plant.Id))
		ch <- prometheus.MustNewConstMetric(m.plantInfo, prometheus.GaugeValue, 1, fmt.Sprintf("%.0f", plant.Id), plant.Name, lastUpdate.UTC().Format(time.RFC3339), fmt.Sprintf("%.0f", maxPower))
		ch <- prometheus.MustNewConstMetric(m.plantPower, prometheus.GaugeValue, plantPower, fmt.Sprintf("%.0f", plant.Id))
	}
}

func (m metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.plantInfo
}
