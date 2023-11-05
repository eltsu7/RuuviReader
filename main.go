package main

import (
	"strings"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// F2:2D:EB:37:8A:D4
// D3:1A:DA:17:E5:C6

func main() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	println("connection handler")
	adapter.SetConnectHandler(connectHandler)
	println("scanning...")
	err := adapter.Scan(handleData)
	println("end")
	must("start scan", err)

}

func handleData(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	var macs = []string{
		"F2:2D:EB:37:8A:D4",
		"D3:1A:DA:17:E5:C6",
	}

	found := false

	if !strings.Contains(device.LocalName(), "Ruuvi") {
		return
	}

	for _, val := range macs {
		if device.Address.String() == val {
			found = true
			break
		}
	}

	if !found {
		return
	}

	println(device.LocalName())

	connectedDevice, err := adapter.Connect(device.Address, bluetooth.ConnectionParams{})
	if err != nil {
		println("Connect error:", err)
	}

	services, err := connectedDevice.DiscoverServices(nil)
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
