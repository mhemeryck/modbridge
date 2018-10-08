package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/mhemeryck/modbridge"
	yaml "gopkg.in/yaml.v2"
)

type empty struct{}

// Current program version info, injected at build time
var version, commit, date string

// printVersionInfo prints the current version info, where the values are injected at build time with goreleaser
func printVersionInfo() {
	fmt.Print("Modbridge\n")
	var info = map[string]string{
		"Version": version,
		"Commit":  commit,
		"Date":    date,
	}
	for key, value := range info {
		fmt.Printf("%s: %s\n", key, value)
	}
}

func main() {
	// Read configuration
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version info and exit")
	var configFile string
	flag.StringVar(&configFile, "filename", "config.yml", "Config file name")
	flag.Parse()

	// Show version and exit
	if showVersion {
		printVersionInfo()
		return
	}

	rawConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading %s\n", configFile)
	}

	var config modbridge.Configuration
	err = yaml.Unmarshal(rawConfig, &config)
	if err != nil {
		log.Fatalf("Error %s reading in config\n", err)
	}
	// modbus client
	modbusClient := modbus.TCPClient(config.ModbusServerURI)

	// MQTT client
	coilMap := config.CoilsMap()
	// Subcribe for each topic: create a callback for all of them
	var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		if coil, ok := coilMap[msg.Topic()]; ok {
			var value uint16
			if string(msg.Payload()) == "ON" {
				value = 0xFF00
			} else {
				value = 0x0000
			}
			_, err := modbusClient.WriteSingleCoil(coil.Address, value)
			if err != nil {
				log.Printf("Error %d writing on MQTT event", err)
			}
		}
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTTBrokerURI)
	opts.SetClientID(config.MQTTClientID)
	opts.OnConnect = func(c mqtt.Client) {
		for slug := range coilMap {
			if token := c.Subscribe(slug, 0, messageHandler); token.Wait() && token.Error() != nil {
				log.Fatal(token.Error())
			}
		}
	}
	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatalf("Can't connect to MQTT host\n")
	}

	coilGroups := config.CoilGroupsList()
	// Attach a copy of the clients to each of the groups
	for k := range coilGroups {
		coilGroups[k].ModbusClient = modbusClient
		coilGroups[k].MQTTClient = mqttClient
	}

	// Continuous infinite polling
	sem := make(chan empty, len(coilGroups))
	for {
		for k := range coilGroups {
			go func(k int) {
				err := coilGroups[k].Update()
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
