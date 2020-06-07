package modbridge

import (
	"sort"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
)

// CoilGroupUpdater exposes the required interfaces for polling CoilGroups
type CoilGroupUpdater interface {
	Update() (err error)
}

// CoilGroup represents an array of coils which have contiguous Addresses
type CoilGroup struct {
	offset       uint16
	coils        []Coil
	ModbusClient modbus.Client
	MQTTClient   mqtt.Client
}

// Update call the modbus group range and update the corresponding coils
func (coilGroup *CoilGroup) Update() (err error) {
	results, err := coilGroup.ModbusClient.ReadCoils(coilGroup.offset, uint16(len(coilGroup.coils)))
	if err != nil {
		return
	}
	for k := range coilGroup.coils {
		numberIndex := k / 8
		bitOffset := uint16(k % 8)
		value := results[numberIndex] & (1 << bitOffset)
		coilGroup.coils[k].Update(value != 0, coilGroup.MQTTClient)
	}
	return
}

// ByAddress implements sorter interface, for sorting an array of coils based on Address
type ByAddress []Coil

func (coils ByAddress) Len() int           { return len(coils) }
func (coils ByAddress) Swap(i, j int)      { coils[i], coils[j] = coils[j], coils[i] }
func (coils ByAddress) Less(i, j int) bool { return coils[i].Address < coils[j].Address }

// GroupCoils groups an array of coils into an array coil groups
func GroupCoils(coils []Coil) []CoilGroup {
	// Sort inputs by Address first
	sort.Sort(ByAddress(coils))

	// Single-length case
	if len(coils) == 1 {
		return []CoilGroup{CoilGroup{offset: coils[0].Address, coils: coils}}
	}

	// Start with the first group by considering it the 1-length case
	groups := GroupCoils(coils[:1])

	// Loop over all in the input and either add them to the existing group, or add a new one
	for _, coil := range coils[1:] {
		groupIndex := len(groups) - 1
		// Compare the curent input Address against the offset + length of the current group
		if coil.Address == groups[groupIndex].offset+uint16(len(groups[groupIndex].coils)) {
			// Add the current input to the current group
			groups[groupIndex].coils = append(groups[groupIndex].coils, coil)
		} else {
			// Start a new group
			groups = append(groups, CoilGroup{offset: coil.Address, coils: []Coil{coil}})
		}
	}
	return groups
}
