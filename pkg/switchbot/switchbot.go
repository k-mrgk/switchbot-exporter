package switchbot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	token      string
	endpoint   string
}

type deviceList struct {
	StatusCode int `json:"statusCode"`
	Body       struct {
		DeviceList []struct {
			DeviceID           string `json:"deviceId"`
			DeviceName         string `json:"deviceName"`
			DeviceType         string `json:"deviceType"`
			EnableCloudService bool   `json:"enableCloudService,omitempty"`
			HubDeviceID        string `json:"hubDeviceId"`
		} `json:"deviceList"`
		InfraredRemoteList []struct {
			DeviceID    string `json:"deviceId"`
			DeviceName  string `json:"deviceName"`
			RemoteType  string `json:"remoteType"`
			HubDeviceID string `json:"hubDeviceId"`
		} `json:"infraredRemoteList"`
	} `json:"body"`
	Message string `json:"message"`
}

type thermometer struct {
	StatusCode int `json:"statusCode"`
	Body       struct {
		DeviceID    string  `json:"deviceId"`
		DeviceType  string  `json:"deviceType"`
		HubDeviceID string  `json:"hubDeviceId"`
		Humidity    int     `json:"humidity"`
		Temperature float64 `json:"temperature"`
	} `json:"body"`
	Message string `json:"message"`
}

const endpoint = "https://api.switch-bot.com"

func NewClient(token string) *Client {

	c := &Client{
		httpClient: http.DefaultClient,

		token:    token,
		endpoint: endpoint,
	}

	return c
}

func (c Client) do(url string) (*http.Response, error) {

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) GetDeviceName(deviceID string) (string, error) {

	url := fmt.Sprintf("%s/v1.0/devices", c.endpoint)
	resp, err := c.do(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP Status Code is %d", resp.StatusCode)
	}

	var b deviceList
	body, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &b); err != nil {
		return "", err
	}

	for _, v := range b.Body.DeviceList {
		if v.DeviceID == deviceID {
			return v.DeviceName, nil
		}
	}
	return "", fmt.Errorf("Device id: %s does not exist", deviceID)
}

func (c Client) GetThermometerValue(deviceID string) (float64, int, error) {

	url := fmt.Sprintf("%s/v1.0/devices/%s/status", c.endpoint, deviceID)
	resp, err := c.do(url)

	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("HTTP Status Code is %d", resp.StatusCode)
	}

	var b thermometer
	body, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &b); err != nil {
		return 0, 0, err
	}

	return b.Body.Temperature, b.Body.Humidity, nil
}
