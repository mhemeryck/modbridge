package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"log"
	"time"
)

func main() {
	// MQTT client
	opts := mqtt.NewClientOptions().AddBroker("tcp://raspberrypi.lan:1883").SetClientID("beppu")
	mqtt_client := mqtt.NewClient(opts)
	token := mqtt_client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// modbus polling
	previous := false
	current := false
	state := "OFF"
	client := modbus.TCPClient("unipi.lan:502")
	for true {
		results, err := client.ReadDiscreteInputs(100, 1)
		if err != nil {
			log.Fatal(err)
		}
		current = results[0] == 1
		if current != previous {
			if current == true {
				if state == "OFF" {
					state = "ON"
				} else {
					state = "OFF"
				}
				mqtt_client.Publish("unipi/digital_input/21", 0, false, state)
			}
			fmt.Printf("%s: %d\n", time.Now().Format(time.UnixDate), results[0])
			previous = current
		}
	}
}
