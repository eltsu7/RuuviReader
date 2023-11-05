package main

import (
	"encoding/json"
	"os"
	"strings"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter
var tags = map[string]string{}

func main() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Read tags from config.json
	file, err := os.ReadFile("config.json")
	if err != nil {
		println(err)
	}
	var configData map[string]interface{}
	json.Unmarshal([]byte(file), &configData)

	configTags, ok := configData["ruuvitags"]
	if !ok {
		panic("Malformed config file")
	}
	println("Tags from config:")
	for k, v := range configTags.(map[string]interface{}) {
		tags[k] = v.(string)
		println("\t", k, v.(string))
	}

	// Start scanning.
	println("connection handler")
	adapter.SetConnectHandler(connectHandler)
	println("scanning...")
	err = adapter.Scan(handleData)
	println("end")
	must("start scan", err)

}

func handleData(adapter *bluetooth.Adapter, scanResult bluetooth.ScanResult) {
	found := false

	if !strings.Contains(scanResult.LocalName(), "Ruuvi") {
		return
	}

	for _, v := range tags {
		if scanResult.Address.String() == v {
			found = true
			break
		}
	}
	if !found {
		return
	}

	device, err := adapter.Connect(scanResult.Address, bluetooth.ConnectionParams{})
	if err != nil {
		println("Connect error:", err)
	}

	services, err := device.DiscoverServices(nil)
	if err != nil {
		println("DiscoverServices error:", err)
	}

	for _, service := range services {
		println("service:", service.UUID().String())

		characteristics, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			println("DiscoverCharacteristics error:", err)
		}

		for _, chara := range characteristics {
			println("enabling notifications..")
			err = chara.EnableNotifications(notification)

			if err != nil {
				println("EnableNotifications error:", err)
			}
		}
	}
}

func notification(buffer []byte) {
	println("NOTIFICATION!")
	println(string(buffer))
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func connectHandler(device bluetooth.Address, connected bool) {
	println(device.String(), connected)
}
