package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/spf13/viper"
	"log"
	"time"
)

func readConfig() {
	viper.AutomaticEnv()
	// defaults
	viper.SetDefault("mqtt_broker_uri", "tcp://raspberrypi.lan:1883")
	viper.SetDefault("mqtt_client_id", "unipusher")
	viper.SetDefault("modbus_server_uri", "unipi.lan:502")
}

func main() {
	// Read configuration
	readConfig()

	// MQTT client
	opts := mqtt.NewClientOptions().AddBroker(viper.GetString("mqtt_broker_uri")).SetClientID(viper.GetString("mqtt_client_id"))
	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// modbus polling
	previous := false
	current := false
	state := "OFF"
	client := modbus.TCPClient(viper.GetString("modbus_server_uri"))
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
				mqttClient.Publish("unipi/digital_input/21", 0, false, state)
			}
			fmt.Printf("%s: %d\n", time.Now().Format(time.UnixDate), results[0])
			previous = current
		}
	}
}
