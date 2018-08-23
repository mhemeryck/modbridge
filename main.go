package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/spf13/viper"
)

// Digital input mode constants
const (
	NO = iota // Normally Open
	NC = iota // Normally Closed
)

var addressMap = map[uint16]string{
	100: "Digital input 2.1",
	101: "Digital input 2.2",
	102: "Digital input 2.3",
	103: "Digital input 2.4",
}

// DigitalInput represents simple digital toggle
type DigitalInput struct {
	address     uint16
	description string
	previous    bool
	current     bool
	lastChange  time.Time
	mode        int
}

// Update poll the modbus server for an updated value
func (input *DigitalInput) Update(modbusClient modbus.Client, mqttClient mqtt.Client) error {
	results, err := modbusClient.ReadDiscreteInputs(input.address, 1)
	input.previous, input.current = input.current, results[0]&0x01 != 0
	if input.current != input.previous {
		input.lastChange = time.Now()
		if (input.mode == NO && input.current && !input.previous) || (input.mode == NC && !input.current && input.previous) {
			fmt.Printf("%s, %s: %v\n", input.Slug(), input.lastChange.Format(time.UnixDate), input.current)
			mqttClient.Publish(input.Slug(), 0, false, "ON")
		}
	}
	return err
}

// Slug representation for pushing out on MQTT
func (input *DigitalInput) Slug() string {
	slug := strings.ToLower(input.description)
	slug = strings.Trim(slug, "_")
	re := regexp.MustCompile("[^a-z0-9-_]")
	return re.ReplaceAllString(slug, "_")
}

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

	// modbus client
	modbusClient := modbus.TCPClient(viper.GetString("modbus_server_uri"))

	// Initialize list of inputs
	var inputs []DigitalInput
	for address, description := range addressMap {
		inputs = append(inputs, DigitalInput{address: address, description: description, previous: false, current: false, mode: NO})
	}
	for {
		for k := range inputs {
			err := inputs[k].Update(modbusClient, mqttClient)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
