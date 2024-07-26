package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseUrl = "https://global.hoymiles.com/platform/api/gateway/"
const LoginPath = "iam/auth_login"
const PvmStationDataPath = "pvm-data/data_count_station_real_data"
const PvmStationsDataPath = "pvm/station_select_by_page"


var apiLog = log.Sugar().Named("api")
var client = http.Client{
	Timeout: 30 * time.Second,
}
var httpLog = log.Sugar().Named("http")
var token = ""

func getPlants() []plantInfo {
	apiLog.Debug("Querying plant information …")

	data := map[string]interface{}{
		"page":      1,
		"page_size": 100,
	}

	var result *plantsData
	res := post(PvmStationsDataPath, data, result)

	return res.Data.List
}

func getPlantData(plantId float64) plantData {
	apiLog.Debug("Querying data for plant ID %0.f …", plantId)

	data := map[string]interface{}{
		"sid": plantId,
	}

	var result *plantData
	res := post(PvmStationDataPath, data, result)

	return res
}

func login(username string, password string) {
	if token != "" {
		apiLog.Debugf("Re-using cached token: %s", token)

		return
	}

	apiLog.Debug("Authenticating with username and password …")

	data := map[string]interface{}{
		"password":  fmt.Sprintf("%x", md5.Sum([]byte(password))),
		"user_name": username,
	}

	var result *loginData
	res := post(LoginPath, data, result)
	token = res.Data.Token

	apiLog.Debugf("Acquired token: %s", token)
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
		apiLog.Fatalf("Error marshalling JSON request body: %s", err.Error())
	}

	apiLog.Debugf("-> %s", jsonBody)
	httpLog.Debugf("-> %s (%d bytes)", path, len(jsonBody))

	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, BaseUrl+path, bodyReader)
	if err != nil {
		apiLog.Fatalf("Error creating HTTP request: %s\n", err.Error())
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := client.Do(req)
	if err != nil {
		httpLog.Errorf("Error sending HTTP request: %s", err)
	}

	httpLog.Debugf("<- HTTP %s (%d bytes)", res.Status, res.ContentLength)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		httpLog.Errorf("Error reading HTTP response body: %s", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		apiLog.Errorf("Error unmarshalling JSON response body: %s", err)
	}

	status := (*result).ApiStatus()
	msg := (*result).ApiMessage()

	apiLog.Debugf("<- API status code %s (%s)", status, msg)

	switch status {
	case "0":
		break
	case "100":
		token = ""
		apiLog.Debug("Token invalidated")

		login(*username, *password)
		return post(path, data, result)
	default:
		apiLog.Errorf("-> API error (%s): %s", status, msg)
	}

	return *result
}
