package modbridge

import (
	"reflect"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mhemeryck/modbridge/mocks"
	"github.com/stretchr/testify/mock"
)

func TestCoilGroupUpdate(t *testing.T) {
	cases := []struct {
		previous bool
		current  bool
		results  []byte
		expected bool
	}{
		{previous: false, current: false, results: []byte{1}, expected: true},
		{previous: false, current: true, results: []byte{0}, expected: false},
	}
	offset := uint16(10)
	Address := uint16(11)
	Slug := "test"
	switchType := NO
	for _, testCase := range cases {
		// Variables
		coils := []Coil{Coil{Address: Address, Slug: Slug, previous: testCase.previous, current: testCase.current, switchType: switchType}}
		ModbusClient := &mocks.ModbusClient{}
		MQTTClient := &mocks.MQTTClient{}
		// Create a coil group
		coilGroup := &CoilGroup{offset: offset, coils: coils, ModbusClient: ModbusClient, MQTTClient: MQTTClient}
		// Prepare test condition
		ModbusClient.On("ReadCoils", coilGroup.offset, uint16(len(coils))).Return(testCase.results, nil)
		MQTTClient.On("Publish", mock.AnythingOfType("string"), byte(0), false, "trigger").Return(&mqtt.PublishToken{})
		// Actual call
		coilGroup.Update()
		// Test case
		if coilGroup.coils[0].current != testCase.expected {
			t.Errorf("Expected current %v but got %v\n", testCase.expected, coilGroup.coils[0].current)
		}
	}
}

func TestGroupCoils(t *testing.T) {
	// Example flat input array of coils
	input := []Coil{Coil{Address: 0}, Coil{Address: 1}, Coil{Address: 10}}
	// Expected results of grouping them
	expected := []CoilGroup{
		CoilGroup{
			offset: 0,
			coils: []Coil{
				Coil{Address: 0}, Coil{Address: 1},
			},
		},
		CoilGroup{
			offset: 10,
			coils: []Coil{
				Coil{Address: 10},
			},
		},
	}
	actual := GroupCoils(input)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Error grouping coils: expected %v, got %v\n", expected, actual)
	}
}
