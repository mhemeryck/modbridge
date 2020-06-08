package modbridge

import (
	"errors"
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
		err      error
	}{
		{previous: false, current: false, results: []byte{1}, expected: true, err: nil},
		{previous: false, current: true, results: []byte{0}, expected: false, err: nil},
		{err: errors.New("bzzt")},
	}
	offset := uint16(10)
	Address := uint16(11)
	Slug := "test"
	switchType := NO
	for _, testCase := range cases {
		// Variables
		coils := []Coil{{Address: Address, Slug: Slug, previous: testCase.previous, current: testCase.current, switchType: switchType}}
		ModbusClient := &mocks.ModbusClient{}
		MQTTClient := &mocks.MQTTClient{}
		// Create a coil group
		coilGroup := &CoilGroup{offset: offset, coils: coils, ModbusClient: ModbusClient, MQTTClient: MQTTClient}
		// Prepare test condition
		ModbusClient.On("ReadCoils", coilGroup.offset, uint16(len(coils))).Return(testCase.results, testCase.err)
		MQTTClient.On("Publish", mock.AnythingOfType("string"), byte(0), false, "trigger").Return(&mqtt.PublishToken{})
		// Actual call
		resultErr := coilGroup.Update()
		if resultErr != testCase.err {
			t.Errorf("Expected error %v but got %v\n", testCase.err, resultErr)
		}
		// Test case, only if no errors occured
		if testCase.err != nil {
			if coilGroup.coils[0].current != testCase.expected {
				t.Errorf("Expected current %v but got %v\n", testCase.expected, coilGroup.coils[0].current)
			}
		}
	}
}

func TestCoilGroupUpdateIndex(t *testing.T) {
	previous := false
	current := false
	results := []byte{0, 1}
	expected := true
	var err error = nil

	offset := uint16(10)
	Address := uint16(19)
	Slug := "test"
	switchType := NO

	// Array of length 9; last index is the test value
	coils := make([]Coil, 9)
	coils[8] = Coil{Address: Address, Slug: Slug, previous: previous, current: current, switchType: switchType}

	ModbusClient := &mocks.ModbusClient{}
	MQTTClient := &mocks.MQTTClient{}
	// Create a coil group
	coilGroup := &CoilGroup{offset: offset, coils: coils, ModbusClient: ModbusClient, MQTTClient: MQTTClient}
	// Prepare test condition
	ModbusClient.On("ReadCoils", coilGroup.offset, uint16(len(coils))).Return(results, err)
	MQTTClient.On("Publish", mock.AnythingOfType("string"), byte(0), false, "trigger").Return(&mqtt.PublishToken{})
	// Actual call
	resultErr := coilGroup.Update()
	if resultErr != nil {
		t.Errorf("Expected error %v but got %v\n", err, resultErr)
	}
	// test the 9th index
	if coilGroup.coils[8].current != expected {
		t.Errorf("Expected current %v but got %v\n", expected, coilGroup.coils[8].current)
	}
}

func TestGroupCoils(t *testing.T) {
	// Example flat input array of coils
	input := []Coil{{Address: 0}, {Address: 10}, {Address: 1}}
	// Expected results of grouping them
	expected := []CoilGroup{
		{
			offset: 0,
			coils: []Coil{
				{Address: 0}, {Address: 1},
			},
		},
		{
			offset: 10,
			coils: []Coil{
				{Address: 10},
			},
		},
	}
	actual := GroupCoils(input)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Error grouping coils: expected %v, got %v\n", expected, actual)
	}
}
