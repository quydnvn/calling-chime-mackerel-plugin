package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    interface{} `json:"message"`
	Data       []struct {
		ID          int    `json:"id"`
		ChannelID   int    `json:"channel_id"`
		SfuSid      string `json:"sfu_sid"`
		Status      int    `json:"status"`
		UUID        string `json:"uuid"`
		Alias       string `json:"alias"`
		MeetingType string `json:"meeting_type"`
		Total       int    `json:"total"`
		TotalActive int    `json:"total_active"`
	} `json:"data"`
}

type ChimePlugin struct {
	Prefix string
}

var data Response

type PluginWithPrefix interface {
	FetchMetrics() (map[string]float64, error)
	GraphDefinition() map[string]mp.Graphs
	MetricKeyPrefix() string
}

func (u ChimePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(u.MetricKeyPrefix())
	return map[string]mp.Graphs{
		"MeetingChimeMetrics": {
			Label: labelPrefix,
			Unit:  mp.UnitFloat,
			Metrics: []mp.Metrics{
				{Name: "Total", Label: "Total"},
				{Name: "Total_Active", Label: "Total_Active"},
				{Name:"Total_Meeting", Label: "Total_Meeting"},
			},
		},
	}
}

func (u ChimePlugin) MetricKeyPrefix() string {
	if u.Prefix == "" {
		u.Prefix = os.Getenv("LABEL")
	}
	return u.Prefix
}

func (u ChimePlugin) FetchMetrics() (map[string]float64, error) {
	metrics := make(map[string]float64)
	err := godotenv.Load("/var/www/calling-chime-mackerel-plugin/.env") // temp .env
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	url := os.Getenv("URL")
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	defer response.Body.Close()
	d, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	err = json.Unmarshal(d, &data)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	total := 0
	totalActive := 0
	for _, k := range data.Data {
		totalActive += k.TotalActive
		total += k.Total
	}
	metrics["Total_Meeting"] = float64(len(data.Data))
	metrics["Total"] = float64(total)
	metrics["Total_Active"] = float64(totalActive)
	return metrics, nil
}

func main() {
	optPrefix := flag.String("metric-key-prefix", os.Getenv("LABEL"), "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()
	u := ChimePlugin{
		Prefix: *optPrefix,
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Tempfile = *optTempfile
	plugin.Run()
}

