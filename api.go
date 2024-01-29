package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const BaseUrl = "https://global.hoymiles.com/platform/api/gateway/"
const LoginPath = "iam/auth_login"
const PvmStationDataPath = "pvm-data/data_count_station_real_data"
const PvmStationsDataPath = "pvm/station_select_by_page"

var client = http.Client{
	Timeout: 30 * time.Second,
}

var token = ""

func getPlantIds() []int {
	data := map[string]interface{}{
		"page":      1,
		"page_size": 100,
	}

	plantsData := post(PvmStationsDataPath, data)

	var plantIds []int
	for _, plantData := range plantsData["list"].([]interface{}) {
		rawId := plantData.(map[string]interface{})["id"].(float64)
		plantIds = append(plantIds, int(rawId))
	}

	return plantIds
}

func getPlantData(plantId int) map[string]interface{} {
	data := map[string]interface{}{
		"sid": plantId,
	}

	return post(PvmStationDataPath, data)
}

func login(username string, password string) {
	data := map[string]interface{}{
		"password":  fmt.Sprintf("%x", md5.Sum([]byte(password))),
		"user_name": username,
	}

	res := post(LoginPath, data)

	token = res["token"].(string)

	log.Printf("Acquired token: %s", token)
}

func post(path string, data map[string]interface{}) map[string]interface{} {
	headers := map[string]string{}
	if path == LoginPath {
		headers["Cookie"] = "hm_token_language=en_us"
	} else {
		headers["Cookie"] = fmt.Sprintf("hm_token=%s", token)
	}

	jsonBody, err := json.Marshal(map[string]interface{}{
		"body": data,
	})
	if err != nil {
		log.Fatalf("Error marshalling JSON request body: %s", err.Error())
	}

	log.Printf("-> %s (%d bytes)", path, len(jsonBody))

	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, BaseUrl+path, bodyReader)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %s\n", err.Error())
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %s\n", err)
	}

	log.Printf("<- %s (%d bytes)", res.Status, res.ContentLength)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading HTTP response body: %s\n", err)
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON response body: %s\n", err)
	}

	status := result["status"].(string)
	if status != "0" {
		msg := result["message"].(string)
		log.Fatalf("-> API error: %s (%s)", msg, status)
	}

	return result["data"].(map[string]interface{})
}
