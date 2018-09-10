package main

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/mhemeryck/modbridge/mocks"
	"github.com/stretchr/testify/mock"
)

func TestRising(t *testing.T) {
	cases := []struct {
		input    DigitalInput
		expected bool
	}{
		{
			input:    DigitalInput{current: true, previous: false},
			expected: true,
		},
		{
			input:    DigitalInput{current: false, previous: true},
			expected: false,
		},
	}

	for _, testCase := range cases {
		if testCase.input.rising() != testCase.expected {
			t.Fail()
		}
	}
}

func TestFalling(t *testing.T) {
	cases := []struct {
		input    DigitalInput
		expected bool
	}{
		{
			input:    DigitalInput{current: false, previous: true},
			expected: true,
		},
		{
			input:    DigitalInput{current: true, previous: false},
			expected: false,
		},
	}

	for _, testCase := range cases {
		if testCase.input.falling() != testCase.expected {
			t.Fail()
		}
	}
}

func TestUpdate(t *testing.T) {
	table := []struct {
		modbusResult     []byte
		modbusError      error
		initialPrevious  bool
		initialCurrent   bool
		expectedPrevious bool
		expectedCurrent  bool
		shouldPublish    bool
		mode             int
	}{
		// Update in case of non-zero value
		{
			modbusResult:     []byte{0x01},
			modbusError:      nil,
			initialPrevious:  false,
			initialCurrent:   false,
			expectedPrevious: false,
			expectedCurrent:  true,
			shouldPublish:    true,
			mode:             NO,
		},
		// No update in case of an error
		{
			modbusResult:     []byte{0x64},
			modbusError:      &modbus.ModbusError{},
			initialPrevious:  false,
			initialCurrent:   false,
			expectedPrevious: false,
			expectedCurrent:  false,
			shouldPublish:    false,
			mode:             NO,
		},
		// Update in case of a low value
		{
			modbusResult:     []byte{0x00},
			modbusError:      nil,
			initialPrevious:  true,
			initialCurrent:   true,
			expectedPrevious: true,
			expectedCurrent:  false,
			shouldPublish:    true,
			mode:             NC,
		},
	}
	for _, testCase := range table {
		// Set up modbus client
		modbusClient := &mocks.ModbusClient{}
		modbusClient.On("ReadDiscreteInputs", mock.AnythingOfType("uint16"), uint16(1)).Return(testCase.modbusResult, testCase.modbusError)
		input := DigitalInput{previous: testCase.initialPrevious, current: testCase.initialCurrent, mode: testCase.mode}
		// Set up mqtt client (conditionally)
		mqttClient := &mocks.MQTTClient{}
		if testCase.shouldPublish {
			mqttClient.On("Publish", mock.AnythingOfType("string"), byte(0), false, "trigger").Return(&mqtt.PublishToken{})
		}
		// Do the actual call
		input.Update(modbusClient, mqttClient)
		// check: modbus calls OK
		modbusClient.AssertExpectations(t)
		// Check: conditionally check for mqtt client called
		if testCase.shouldPublish {
			mqttClient.AssertExpectations(t)
		}
		// Check previous value
		if input.previous != testCase.expectedPrevious {
			t.Errorf("Expected previous %v but got %v\n", testCase.expectedPrevious, input.previous)
		}
		// Check current value
		if input.current != testCase.expectedCurrent {
			t.Errorf("Expected current %v but got %v\n", testCase.expectedCurrent, input.current)
		}
	}
}
