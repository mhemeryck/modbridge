package modbridge

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SwitchType indicates how to interpret the switch
type SwitchType int

// Switch type constants
const (
	NO SwitchType = iota // Normally Open
	NC                   // Normally Closed
)

// CoilUpdater represents a coil to be polled
type CoilUpdater interface {
	Update()
}

// Coil represents the state we keep about a modbus coil
type Coil struct {
	Address    uint16
	Slug       string
	previous   bool
	current    bool
	switchType SwitchType
}

// rising checks whether the value switched from false to true
func (coil *Coil) rising() bool {
	return coil.current && !coil.previous
}

// falling checks whether the value switched from true to false
func (coil *Coil) falling() bool {
	return !coil.current && coil.previous
}

// Update handles checking a new value against the current and previous retained state we have for a coil
func (coil *Coil) Update(value bool, mqttClient mqtt.Client) {
	coil.previous, coil.current = coil.current, value
	if coil.current != coil.previous {
		if (coil.switchType == NO && coil.rising()) || (coil.switchType == NC && coil.falling()) {
			log.Printf("%s  -  trigger for %s", time.Now().Format(time.RFC3339), coil.Slug)
			mqttClient.Publish(coil.Slug, 0, false, "trigger")
		}
	}
}
