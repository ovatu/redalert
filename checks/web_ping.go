package checks

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jonog/redalert/utils"
)

func init() {
	Register("web-ping", NewWebPinger)
}

type WebPinger struct {
	Address string
	log     *log.Logger
}

var WebPingerMetrics = map[string]MetricInfo{
	"latency": MetricInfo{
		Unit: "ms",
	},
}

type WebPingerConfig struct {
	Address string `json:"address"`
}

var NewWebPinger = func(config Config, logger *log.Logger) (Checker, error) {
	var webPingerConfig WebPingerConfig
	err := json.Unmarshal([]byte(config.Config), &webPingerConfig)
	if err != nil {
		return nil, err
	}
	if webPingerConfig.Address == "" {
		return nil, errors.New("web-ping: address to ping cannot be blank")
	}
	return Checker(&WebPinger{webPingerConfig.Address, logger}), nil
}

var GlobalClient = http.Client{
	Timeout: time.Duration(10 * time.Second),
}

func (wp *WebPinger) Check() (Metrics, error) {

	metrics, err := wp.ping()
	if err != nil {
		// if the initial ping fails, retry after 5 seconds
		// the retry is to avoid noise from intermittent network/connection issues
		time.Sleep(5 * time.Second)
		return wp.ping()
	}

	return metrics, nil
}

func (wp *WebPinger) ping() (Metrics, error) {

	metrics := Metrics(make(map[string]*float64))

	latency := float64(0)
	defer func() {
		metrics["latency"] = &latency
	}()

	startTime := time.Now()
	wp.log.Println("GET", wp.Address)

	req, err := http.NewRequest("GET", wp.Address, nil)
	if err != nil {
		return metrics, errors.New("web-ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	resp, err := GlobalClient.Do(req)
	if err != nil {
		return metrics, errors.New("web-ping: failed client.Do " + err.Error())
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latencyCalc := endTime.Sub(startTime)
	latency = float64(latencyCalc.Seconds() * 1e3)

	wp.log.Println("Latency", utils.White, latency, utils.Reset)

	if err != nil {
		return metrics, errors.New("web-ping: failed reading body " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return metrics, errors.New("web-ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}

	return metrics, nil
}

func (wp *WebPinger) MetricInfo(metric string) MetricInfo {
	return WebPingerMetrics[metric]
}

func (wp *WebPinger) MessageContext() string {
	return wp.Address
}
