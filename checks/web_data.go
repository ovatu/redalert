package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ovatu/redalert/data"
	"github.com/ovatu/redalert/utils"

	"github.com/jmoiron/jsonq"
)

func init() {
	Register("web-data", NewWebData)
}

type WebData struct {
	Config          Config
	WebDataConfig WebDataConfig
	log             *log.Logger
}

var WebDataMetrics = map[string]MetricInfo{
}

var WebDataConfigMetrics = map[string]WebMetricConfig{}

type WebDataConfig struct {
	Address string            `json:"address"`
	Headers map[string]string `json:"headers"`
	Metrics []WebMetricConfig `json:"metrics"`
}

type WebMetricConfig struct {
	Name string 		`json:"name"`
	Unit string			`json:"unit"`
	Identifier string 	`json:"identifier"`
}

var NewWebData = func(config Config, logger *log.Logger) (Checker, error) {
	var WebDataConfig WebDataConfig
	err := json.Unmarshal([]byte(config.Config), &WebDataConfig)
	if err != nil {
		return nil, err
	}
	if WebDataConfig.Address == "" {
		return nil, errors.New("web-data: address to ping cannot be blank")
	}

	for _, metricConfig := range WebDataConfig.Metrics {
		fmt.Println(metricConfig.Name)
		fmt.Println(metricConfig.Unit)
		WebDataMetrics[metricConfig.Name] = MetricInfo{
			Unit: metricConfig.Unit,
		}

		WebDataConfigMetrics[metricConfig.Name] = metricConfig
	}

	return Checker(&WebData{config, WebDataConfig, logger}), nil
}

func (wp *WebData) Check() (data.CheckResponse, error) {
	metadata := make(map[string]string)
	metrics, b, statusCode, err := wp.ping()
	if err != nil {
		// if the initial ping fails, retry after 5 seconds
		// the retry is to avoid noise from intermittent network/connection issues
		time.Sleep(5 * time.Second)
		var secondMetrics map[string]*float64
		var secondStatusCode int
		var secondB []byte
		secondMetrics, secondB, secondStatusCode, err = wp.ping()
		metadata["status_code"] = strconv.Itoa(secondStatusCode)
		return data.CheckResponse{Metrics: secondMetrics, Metadata: metadata, Response: secondB}, err
	}
	metadata["status_code"] = strconv.Itoa(statusCode)
	return data.CheckResponse{Metrics: metrics, Metadata: metadata, Response: b}, nil
}

func (wp *WebData) ping() (data.Metrics, []byte, int, error) {

	metrics := data.Metrics(make(map[string]*float64))
	var b []byte

	latency := float64(0)

	startTime := time.Now()
	if wp.Config.VerboseLogging != nil && *wp.Config.VerboseLogging {
		wp.log.Println("GET", wp.WebDataConfig.Address)
	}

	req, err := http.NewRequest("GET", wp.WebDataConfig.Address, nil)
	if err != nil {
		return metrics, b, 0, errors.New("web-ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	for k, v := range wp.WebDataConfig.Headers {
		req.Header.Add(k, v)
	}

	resp, err := GlobalClient.Do(req)
	if err != nil {
		return metrics, b, 0, errors.New("web-ping: failed client.Do " + err.Error())
	}

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latencyCalc := endTime.Sub(startTime)
	latency = float64(latencyCalc.Seconds() * 1e3)

	data := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.Decode(&data)

	fmt.Println(data)

	for _, metricConfig := range WebDataConfigMetrics {
		fmt.Println(metricConfig.Identifier)
		fmt.Println(strings.Split(metricConfig.Identifier, "."))

		jq := jsonq.NewQuery(data)
		floatField, err := jq.Float(strings.Split(metricConfig.Identifier, ".")...)

		if err == nil {
			metrics[metricConfig.Name] = &floatField
			continue
		}

		intField, err := jq.Int(strings.Split(metricConfig.Identifier, ".")...)
		if err == nil {
			floatField := float64(intField)
			metrics[metricConfig.Name] = &floatField
			continue
		}

		stringField, err := jq.String(strings.Split(metricConfig.Identifier, ".")...)

		if err == nil {
			f, err := strconv.ParseFloat(strings.TrimSpace(stringField), 64)
			if err != nil {
				return metrics, b, 0, errors.New("command: error while parsing number: " + err.Error())
			}
			metrics[metricConfig.Name] = &f
		} else {
			return metrics, b, 0, errors.New("json: invalid identifier")
		}
	}



	if wp.Config.VerboseLogging != nil && *wp.Config.VerboseLogging {
		wp.log.Println("Latency", utils.White, latency, utils.Reset)
	}

	if err != nil {
		return metrics, b, resp.StatusCode, errors.New("web-ping: failed reading body " + err.Error())
	}

	return metrics, b, resp.StatusCode, nil
}

func (wp *WebData) MetricInfo(metric string) MetricInfo {
	return WebDataMetrics[metric]
}

func (wp *WebData) MessageContext() string {
	return wp.WebDataConfig.Address
}
