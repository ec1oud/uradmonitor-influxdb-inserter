package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"

// {"data":{ "id":"41000008","type":"4","detector":"SBM20","voltage":379,"cpm":31,"temperature":11.00,"pressure":99815,"uptime": 480}}
type URadMonitorData struct {
    Id string
    Type string
    Detector string
    Voltage int
    Cpm int
    Temperature float64
    Pressure int
    Uptime int
}

type URadMonitorDataData struct {
	Data URadMonitorData
}

func main() {
	resp, err := http.Get("http://10.0.0.186/j")
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
	var data URadMonitorDataData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(data.Data)
}
