package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

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

// NewTLSConfig generates a TLS config instance for use with the MQTT setup
func NewTLSConfig(caFile string, insecure bool) *tls.Config {
	// Read the ceritifcates from the system, continue with empty pool in case of failure
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read the local file from the supplied path
	certs, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to append %q to RootCAs: %v", caFile, err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	return &tls.Config{
		InsecureSkipVerify: insecure,
		RootCAs:            rootCAs,
	}
}

func main() {
	// Read configuration
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version info and exit")
	var configFile string
	flag.StringVar(&configFile, "filename", "config.yml", "Config file name")
	var pollingInterval int
	flag.IntVar(&pollingInterval, "polling_interval", 20, "Polling interval for one coil group in millis")
	var caFile string
	flag.StringVar(&caFile, "cafile", "", "CA certificate used for MQTT TLS setup")
	var insecure bool
	flag.BoolVar(&insecure, "insecure", false, "Flag to control MQTT host TLS host name check")
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
	if caFile != "" {
		tlsConfig := NewTLSConfig(caFile, insecure)
		opts.SetTLSConfig(tlsConfig)
	}
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

	ticker := time.NewTicker(time.Millisecond * time.Duration(pollingInterval)).C
	// Continuous infinite polling
	k := 0
	for {
		select {
		case <-ticker:
			err := coilGroups[k].Update()
			if err != nil {
				log.Fatal(err)
			}
			k = (k + 1) % len(coilGroups)
		}
	}
}
