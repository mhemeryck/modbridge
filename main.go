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
	104: "Digital input 2.5",
	105: "Digital input 2.6",
	106: "Digital input 2.7",
	107: "Digital input 2.8",
	108: "Digital input 2.9",
	109: "Digital input 2.10",
	110: "Digital input 2.11",
	111: "Digital input 2.12",
	112: "Digital input 2.13",
	113: "Digital input 2.14",
	114: "Digital input 2.15",
	115: "Digital input 2.16",
	116: "Digital input 2.17",
	117: "Digital input 2.18",
	118: "Digital input 2.19",
	119: "Digital input 2.20",
	120: "Digital input 2.21",
	121: "Digital input 2.22",
	122: "Digital input 2.23",
	123: "Digital input 2.24",
	124: "Digital input 2.25",
	125: "Digital input 2.26",
	126: "Digital input 2.27",
	127: "Digital input 2.28",
	128: "Digital input 2.29",
	129: "Digital input 2.30",
}

// Input represents any modbus input
type Input interface {
	Update() error
	Slug() string
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

func (input *DigitalInput) rising() bool {
	return input.current && !input.previous
}

func (input *DigitalInput) falling() bool {
	return !input.current && input.previous
}

// Update poll the modbus server for an updated value
func (input *DigitalInput) Update(modbusClient modbus.Client, mqttClient mqtt.Client) error {
	results, err := modbusClient.ReadDiscreteInputs(input.address, 1)
	if err != nil {
		return err
	}
	input.previous, input.current = input.current, results[0]&0x01 != 0
	if input.current != input.previous {
		input.lastChange = time.Now()
		if (input.mode == NO && input.rising()) || (input.mode == NC && input.falling()) {
			mqttClient.Publish(input.Slug(), 0, false, "trigger")
			fmt.Printf("%s, %s: %v\n", input.Slug(), input.lastChange.Format(time.UnixDate), input.current)
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
	viper.SetDefault("mqtt_client_id", "modbridge")
	viper.SetDefault("modbus_server_uri", "unipi.lan:502")
}

type empty struct{}

func main() {
	// Read configuration
	readConfig()

	sem := make(chan empty, 2)
	// MQTT client
	opts := mqtt.NewClientOptions().AddBroker(viper.GetString("mqtt_broker_uri")).SetClientID(viper.GetString("mqtt_client_id"))
	mqttClient := mqtt.NewClient(opts)
	go func() {
		for {
			token := mqttClient.Connect()
			if token.Wait() && token.Error() != nil {
				fmt.Printf("Can't connect to MQTT host, retrying ...\n")
			} else {
				sem <- empty{}
				break
			}
		}
	}()

	// Modbus client
	handler := modbus.NewTCPClientHandler(viper.GetString("modbus_server_uri"))
	go func() {
		for {
			err := handler.Connect()
			if err != nil {
				fmt.Print("Can't connect to Modbus server, retrying ...\n")
			} else {
				sem <- empty{}
				break
			}
		}
	}()
	defer handler.Close()
	modbusClient := modbus.NewClient(handler)

	// Wait for concurrent tasks to finish
	for i := 0; i < cap(sem); i++ {
		<-sem
	}

	// Initialize list of inputs
	var inputs []DigitalInput
	for address, description := range addressMap {
		inputs = append(inputs, DigitalInput{address: address, description: description, previous: false, current: false, mode: NO})
	}

	// Infinite loop
	sem = make(chan empty, 8)
	for {
		for k := range inputs {
			go func(k int) {
				err := inputs[k].Update(modbusClient, mqttClient)
				if err != nil {
					log.Fatal(err)
				}
				sem <- empty{}
			}(k)
		}
		for i := 0; i < cap(sem); i++ {
			<-sem
		}
	}
}
