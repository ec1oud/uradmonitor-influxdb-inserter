package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"log"
	influx "github.com/influxdata/influxdb/client/v2"
)

// {"data":{ "id":"41000008","type":"4","detector":"SBM20","voltage":379,"cpm":31,"temperature":11.00,"pressure":99815,"uptime": 480}}
// {"data":{ "id":"820000ED","type":"8","detector":"SI29BG","cpm":19,"voltage":381,"temperature":-0.74,"humidity":58.50,"pressure":101081,"voc":277472,"co2":353,"noise":23.67,"ch2o":0.00,"pm25":3,"uptime": 121921}}
type URadMonitorData struct {
    Id string
    Type string
    Detector string
    Voltage int
    Cpm int
    Voc int
    Co2 int
    Noise float64
    Pm25 int
    Ch20 float64
    Temperature float64
    Humidity float64
    Pressure int
    Uptime int
}

type URadMonitorDataData struct {
	Data URadMonitorData
}

func main() {
	resp, err := http.Get("http://10.42.101.249/j")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
	var dataData URadMonitorDataData
	err = json.Unmarshal(body, &dataData)
	data := dataData.Data
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(data)
	influxUrl := "http://localhost:8086"
	client, err := influx.NewHTTPClient(influx.HTTPConfig{Addr: influxUrl})
	if err != nil {
		fmt.Println(err)
		return
	}
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "weather",
	})
	point, err := influx.NewPoint(
		"uradmonitor",
		map[string]string{
			"stationId": data.Id,
		},
		map[string]interface{}{
			"voltage": data.Voltage,
			"cpm": data.Cpm,
			"temperature": data.Temperature,
			"humidity": data.Humidity,
			"pressure": data.Pressure,
			"voc": data.Voc,
			"co2": data.Co2,
			"noise": data.Noise,
			"pm25": data.Pm25,
			"ch20": data.Ch20,
		},
		time.Now(),
	)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	bp.AddPoint(point)
	err = client.Write(bp)
	if err != nil {
		log.Fatal(err)
		return
	}
}
