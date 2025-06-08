package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap/zapcore"
)

const BaseUrl = "https://neapi.hoymiles.com/"
const LoginPath = "iam/pub/0/auth/login"
const PvmStationDataPath = "pvm-data/api/0/station/data/count_station_real_data"
const PvmStationsDataPath = "pvm/api/0/station/select_by_page"

var apiLog = log.Sugar().Named("api")
var client = http.Client{
	Timeout: 30 * time.Second,
}
var httpLog = log.Sugar().Named("http")
var token = ""

func getPlants() []plantInfo {
	apiLog.Info("Querying plant information …")

	data := map[string]interface{}{
		"page":      1,
		"page_size": 100,
	}

	var result *plantsData
	res, _ := post(PvmStationsDataPath, data, result)

	return res.Data.List
}

func getPlantData(plantId float64) plantData {
	apiLog.Infof("Querying data for plant ID %0.f …", plantId)

	data := map[string]interface{}{
		"sid": plantId,
	}

	var result *plantData
	res, _ := post(PvmStationDataPath, data, result)

	return *res
}

func login(username string, password string) error {
	if token != "" {
		apiLog.Debugf("Re-using cached token: %s", token)

		return nil
	}

	apiLog.Info("Authenticating with username and password …")

	sha256Password := sha256.Sum256([]byte(password))
	encPassword := fmt.Sprintf("%x", md5.Sum([]byte(password))) + "." + base64.StdEncoding.EncodeToString(sha256Password[:])
	data := map[string]interface{}{
		"password":  encPassword,
		"user_name": username,
	}

	var result *loginData
	res, err := post(LoginPath, data, result)
	if err != nil {
		apiLog.Errorf("Login failed for user '%s': %s", username, err)
		return err
	}

	loginToken, ok := res.Data.(map[string]interface{})
	if !ok {
		apiLog.Errorf("Failed to parse login token for user '%s'.", username)
		return fmt.Errorf("failed to parse login token for user '%s'", username)
	}

	token, ok = loginToken["token"].(string)
	if !ok {
		apiLog.Errorf("Failed to parse login token for user '%s'.", username)
		return fmt.Errorf("failed to parse login token for user '%s'", username)
	}

	apiLog.Debugf("Acquired token: %s", token)

	return nil
}

func post[T response](path string, data map[string]interface{}, result *T) (*T, error) {
	headers := map[string]string{}
	if path != LoginPath {
		headers["Authorization"] = token
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		apiLog.Fatalf("Error marshalling JSON request body: %s", err.Error())
	}

	if apiLog.Level().Enabled(zapcore.DebugLevel) {
		redactedData := data
		if _, exists := redactedData["password"]; exists {
			redactedData["password"] = "[REDACTED]"
		}
		redactedJsonBody, _ := json.Marshal(redactedData)

		apiLog.Debugf("-> %s", redactedJsonBody)
	}

	httpLog.Debugf("-> %s (%d bytes)", path, len(jsonBody))

	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, BaseUrl+path, bodyReader)
	if err != nil {
		apiLog.Fatalf("Error creating HTTP request: %s\n", err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := client.Do(req)
	if err != nil {
		httpLog.Errorf("Error sending HTTP request: %s", err)
		return nil, err
	}

	if res.StatusCode == http.StatusNotFound {
		httpLog.Errorf("API endpoint not found: %s", req.URL)
		return nil, fmt.Errorf("API endpoint not found: %s", req.URL)
	}

	httpLog.Debugf("<- HTTP %s (%d bytes)", res.Status, res.ContentLength)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		httpLog.Errorf("Error reading HTTP response body: %s", err)
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		apiLog.Errorf("Error unmarshalling JSON response body: %s", err)
		return nil, err
	}

	status := (*result).ApiStatus()
	msg := (*result).ApiMessage()

	apiLog.Debugf("<- API status code %s (%s)", status, msg)

	switch status {
	case "0":
		break
	case "1":
		return nil, fmt.Errorf("%s", msg)
	case "100":
		token = ""
		apiLog.Debug("Token invalidated")

		login(*username, *password)
		return post(path, data, result)
	default:
		apiLog.Errorf("-> API error (%s): %s", status, msg)
	}

	return result, nil
}
