package main

type response interface {
	deviceData | loginData | plantData | plantsData

	ApiMessage() string
	ApiStatus() string
}

type baseResponse struct {
	Status       string
	Message      string
	SystemNotice *string
}

func (r baseResponse) ApiMessage() string {
	return r.Message
}

func (r baseResponse) ApiStatus() string {
	return r.Status
}

type deviceData struct {
	baseResponse
}

type loginData struct {
	baseResponse
	Data struct {
		Token string
	}
}

type plantData struct {
	baseResponse
	Data struct {
		LastDataTime string `json:"last_data_time"`
		RealPower    string `json:"real_power"`
	}
}

type plantInfo struct {
	Id   float64
	Name string
}

type plantsData struct {
	baseResponse
	Data struct {
		List []plantInfo
	}
}
