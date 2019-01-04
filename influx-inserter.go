package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
    "strconv"
	"time"
	"log"
	influx "github.com/influxdata/influxdb/client/v2"
)

const influxUrl string = "http://localhost:8086"
const uRadMonitorUrl string = "http://10.42.101.249/j"
const uRadMonitorTemperatureCorrection float64 = 3
const oneWireSensorDir string = "/run/owfs/"
const oneWireSensorId string = "26.A97D5A000000"

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

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func pollAndInsertFromURadMonitor() {
	resp, err := http.Get(uRadMonitorUrl)
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
	fmt.Println(string(body))
	var dataData URadMonitorDataData
	err = json.Unmarshal(body, &dataData)
	data := dataData.Data
	if err != nil {
		fmt.Println(err)
		return
	}
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
			"temperature": data.Temperature - uRadMonitorTemperatureCorrection,
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

func readOneWireSensor(devId string, sensor string) float64 {
	var b bytes.Buffer
	b.WriteString(oneWireSensorDir)
	b.WriteString(devId)
	b.WriteString("/")
	b.WriteString(sensor)
	data, err := ioutil.ReadFile(b.String())
    check(err)
	ret, err := strconv.ParseFloat(string(data), 64)
	check(err)
	return ret
}

func pollAndInsertFromDatanab() {
	client, err := influx.NewHTTPClient(influx.HTTPConfig{Addr: influxUrl})
	if err != nil {
		fmt.Println(err)
		return
	}
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "weather",
	})
	point, err := influx.NewPoint(
		"onewire",
		map[string]string{
			"stationId": oneWireSensorId,
		},
		map[string]interface{}{
			"temperature": readOneWireSensor(oneWireSensorId, "temperature"),
			"humidity": readOneWireSensor(oneWireSensorId, "humidity"),
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

func main() {
	pollAndInsertFromURadMonitor()
	pollAndInsertFromDatanab()
}
