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

func getPlants() []plantInfo {
	data := map[string]interface{}{
		"page":      1,
		"page_size": 100,
	}

	var result *plantsData
	res := post(PvmStationsDataPath, data, result)

	return res.Data.List
}

func getPlantData(plantId float64) plantData {
	data := map[string]interface{}{
		"sid": plantId,
	}

	var result *plantData
	res := post(PvmStationDataPath, data, result)

	return res
}

func login(username string, password string) {
	if token != "" {
		log.Printf("Re-using cached token: %s", token)

		return
	}

	data := map[string]interface{}{
		"password":  fmt.Sprintf("%x", md5.Sum([]byte(password))),
		"user_name": username,
	}

	var result *loginData
	res := post(LoginPath, data, result)
	token = res.Data.Token

	log.Printf("Acquired token: %s", token)
}

func post[T response](path string, data map[string]interface{}, result *T) T {
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

	log.Printf("<- HTTP %s (%d bytes)", res.Status, res.ContentLength)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading HTTP response body: %s\n", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON response body: %s\n", err)
	}

	status := (*result).ApiStatus()
	msg := (*result).ApiMessage()

	log.Printf("<- API %s (%s)", status, msg)

	switch status {
	case "0":
		break
	case "100":
		token = ""
		log.Printf("Token invalidated")

		login(*username, *password)
		return post(path, data, result)
	default:
		log.Fatalf("-> API error: %s (%s)", msg, status)
	}

	return *result
}
