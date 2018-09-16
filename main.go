package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/spf13/viper"
)

// Digital input mode constants
const (
	NO = iota // Normally Open
	NC        // Normally Closed
)

// Input represents any modbus input
type Input interface {
	Update() error
}

// DigitalInput represents simple digital toggle
type DigitalInput struct {
	address    uint16
	slug       string
	previous   bool
	current    bool
	lastChange time.Time
	mode       int
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
			mqttClient.Publish(input.slug, 0, false, "trigger")
			fmt.Printf("%s, %s: %v\n", input.slug, input.lastChange.Format(time.UnixDate), input.current)
		}
	}
	return err
}

type empty struct{}

func main() {
	// Read configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error %s reading in file", err))
	}

	var config Configuration
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Error %s reading in config", err))
	}

	sem := make(chan empty, 2)
	// MQTT client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTTBrokerURI)
	opts.SetClientID(config.MQTTClientID)
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
	handler := modbus.NewTCPClientHandler(config.ModbusServerURI)
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
	inputMap := make(map[string]DigitalInput)
	for _, coil := range config.Coils {
		input := DigitalInput{address: coil.Address, slug: coil.Slug, previous: false, current: false, mode: NO}
		// Skip in case we only use it to write back
		if coil.Mode != "W" {
			inputs = append(inputs, input)
		}
		inputMap[coil.Slug] = input
	}

	// Subcribe for each topic: create a callback for all of them
	var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		if input, ok := inputMap[msg.Topic()]; ok {
			var value uint16
			if string(msg.Payload()) == "ON" {
				value = 0xFF00
			} else {
				value = 0x0000
			}
			_, err := modbusClient.WriteSingleCoil(input.address, value)
			if err != nil {
				fmt.Printf("Error %d writing on MQTT event", err)
			}
		}
	}

	// Infinite loop
	sem = make(chan empty, 8)
	for {
		// Async polling modbus to MQTT
		for k := range inputs {
			go func(k int) {
				err := inputs[k].Update(modbusClient, mqttClient)
				if err != nil {
					log.Fatal(err)
				}
				sem <- empty{}
			}(k)
		}
		// Sync polling MQTT to modbus
		for slug := range inputMap {
			if token := mqttClient.Subscribe(slug, 0, messageHandler); token.Wait() && token.Error() != nil {
				log.Fatal(err)
			}
		}
		for i := 0; i < cap(sem); i++ {
			<-sem
		}
	}
}
