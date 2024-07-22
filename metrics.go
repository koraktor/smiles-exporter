package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	plantEnergyTotal *prometheus.Desc
	plantInfo        *prometheus.Desc
	plantLastUpdate  *prometheus.Desc
	plantPower       *prometheus.Desc
}

var timeZoneRegex = regexp.MustCompile(`UTC([+-])(\d{2})`)

var collectorLog = log.Sugar().Named("collector")

func newMetrics() metrics {
	return metrics{
		plantEnergyTotal: prometheus.NewDesc(
			"smiles_plant_energy_total",
			"Total energy produced by a monitored plant",
			[]string{"plant_id"}, nil),

		plantInfo: prometheus.NewDesc(
			"smiles_plant_info",
			"Basic information about the monitored plant",
			[]string{"plant_id", "name", "max_power"}, nil),

		plantLastUpdate: prometheus.NewDesc(
			"smiles_plant_last_update",
			"Last time a plant sent a status update to S-Miles Cloud",
			[]string{"plant_id"}, nil),

		plantPower: prometheus.NewDesc(
			"smiles_plant_power",
			"Current power produced by the monitored plant",
			[]string{"plant_id"}, nil),
	}
}

func (m metrics) Collect(ch chan<- prometheus.Metric) {
	collectorLog.Debug("Collecting metrics …")

	login(*username, *password)
	plantInfo := getPlants()

	for _, plant := range plantInfo {
		collectorLog.Debugf("Building metrics for plant ID %0.f …", plant.Id)

		plantData := getPlantData(plant.Id)

		energyToday, _ := strconv.ParseFloat(plantData.Data.EnergyToday, 64)
		energyTotal, _ := strconv.ParseFloat(plantData.Data.EnergyTotal, 64)
		maxPower, _ := strconv.ParseFloat(plantData.Data.MaxPower, 64)
		maxPower *= 1000

		timeZoneOffset := 0
		timeZoneResult := timeZoneRegex.FindAllStringSubmatch(plant.TimeZone, 1)
		if len(timeZoneResult) > 0 {
			timeZone := timeZoneResult[0]
			timeZoneNegative := timeZone[1] == "-"
			timeZoneOffset, _ := strconv.ParseInt(timeZone[2], 10, 32)
			if timeZoneNegative {
				timeZoneOffset *= -1
			}
		}

		location := time.FixedZone(plant.TimeZone, int(timeZoneOffset)*3600)
		lastUpdate, _ := time.ParseInLocation(time.DateTime, plantData.Data.LastDataTime, location)

		plantPower, _ := strconv.ParseFloat(plantData.Data.RealPower, 64)

		ch <- prometheus.MustNewConstMetric(m.plantEnergyTotal, prometheus.CounterValue, energyTotal+energyToday, fmt.Sprintf("%.0f", plant.Id))
		ch <- prometheus.MustNewConstMetric(m.plantInfo, prometheus.GaugeValue, 1, fmt.Sprintf("%.0f", plant.Id), plant.Name, fmt.Sprintf("%.0f", maxPower))
		ch <- prometheus.MustNewConstMetric(m.plantLastUpdate, prometheus.CounterValue, float64(lastUpdate.UTC().Unix()), fmt.Sprintf("%.0f", plant.Id))
		ch <- prometheus.MustNewConstMetric(m.plantPower, prometheus.GaugeValue, plantPower, fmt.Sprintf("%.0f", plant.Id))
	}
}

func (m metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.plantInfo
}
