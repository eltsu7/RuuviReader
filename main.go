package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var wantedTags = map[string]string{}
var connectedTags = []string{}

func main() {
	// Enable BLE interface.
	var adapter = bluetooth.DefaultAdapter
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
	println("Wanted tags from config:")
	for k, v := range configTags.(map[string]interface{}) {
		wantedTags[k] = v.(string)
		println("\t", k, v.(string))
	}

	// Start scanning.
	adapter.SetConnectHandler(connectHandler)
	for true {
		if !allTagsConnected() {
			println("All tags are not connected, starting to scan...")
			err = adapter.Scan(handleData)
			checkError(err, "Scanning error")
		}
		time.Sleep(1 * time.Second)
	}
	println("end")
}

func handleData(adapter *bluetooth.Adapter, scanResult bluetooth.ScanResult) {
	mac := scanResult.Address.String()

	// Filter out unwanted bluetooth devices
	found := false
	for _, tagMac := range wantedTags {
		if mac == tagMac {
			found = true
			break
		}
	}
	if !found {
		return
	}

	// Filter out all but ruuvi tags. Ruuvi sends 2 ble messages for some reason,
	// the one with LocalName "Ruuvi ..." is the correct one.
	if !strings.Contains(scanResult.LocalName(), "Ruuvi") {
		return
	}

	// Filter our already connected tags
	for _, tagMac := range connectedTags {
		if tagMac == mac {
			return
		}
	}

	device, err := adapter.Connect(scanResult.Address, bluetooth.ConnectionParams{})
	checkError(err, "Connection error")

	services, err := device.DiscoverServices(nil)
	checkError(err, "DiscoverServices error")

	for _, service := range services {
		characteristics, err := service.DiscoverCharacteristics(nil)
		checkError(err, "DiscoverCharacteristics error")

		for _, chara := range characteristics {

			alreadyConnected := false
			for _, tagMac := range connectedTags {
				if tagMac == mac {
					alreadyConnected = true
					break
				}
			}
			if alreadyConnected {
				continue
			}

			if chara.Properties() == 16 {
				println("enabling notifications for", scanResult.LocalName())

				err = chara.EnableNotifications(notification)
				checkError(err, "EnableNotifications error")
				if err != nil {
					continue
				}
				connectedTags = append(connectedTags, mac)
			}
		}
	}

	if allTagsConnected() {
		println("All wanted tags connected! Stopping scan...")
		adapter.StopScan()
	}
}

func notification(buffer []byte) {

	// for _, b := range tempBytes {
	// 	// print(fmt.Sprintf("%2x", b))
	// 	fmt.Printf("%08b", b)
	// }
	// println()

	tempBytes := buffer[1:3]
	temp := float32(big.NewInt(0).SetBytes(tempBytes).Int64()) * 0.005

	humidityBytes := buffer[3:5]
	humidity := float32(big.NewInt(0).SetBytes(humidityBytes).Int64()) * 0.0025

	msg := "New data:\n"
	msg += "\tTemp: " + fmt.Sprint(temp) + " C\n"
	msg += "\tHumidity: " + fmt.Sprint(humidity) + " %\n"

	print(msg)

	// hexString := hex.EncodeToString(buffer)
	// println(len(buffer))
	// bitString := ""
	// for _, b := range buffer {
	// 	bitString = bitString + fmt.Sprintf("%08b", b)
	// }
	// println("NOTIFICATION: " + hexString + " " + bitString)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func allTagsConnected() bool {
	for _, wantedTag := range wantedTags {
		found := false
		for _, connectedTag := range connectedTags {
			if connectedTag == wantedTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func connectHandler(device bluetooth.Address, connected bool) {
	println("ASDASD")
	connectedString := ""
	if connected {
		connectedString = "CONNECTED"
	} else {
		connectedString = "DISCONNECTED"
	}
	println(device.String(), connectedString)
}

func checkError(err error, message string) {
	if err != nil {
		println("ERROR", err, message)
	}
}
