package modbridge

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mhemeryck/modbridge/mocks"
	"github.com/stretchr/testify/mock"
)

func TestCoilRising(t *testing.T) {
	cases := []struct {
		coil     Coil
		expected bool
	}{
		{
			coil:     Coil{current: true, previous: false},
			expected: true,
		},
		{
			coil:     Coil{current: false, previous: true},
			expected: false,
		},
	}

	for _, testCase := range cases {
		if testCase.coil.rising() != testCase.expected {
			t.Fail()
		}
	}
}

func TestCoilFalling(t *testing.T) {
	cases := []struct {
		coil     Coil
		expected bool
	}{
		{
			coil:     Coil{current: false, previous: true},
			expected: true,
		},
		{
			coil:     Coil{current: true, previous: false},
			expected: false,
		},
	}

	for _, testCase := range cases {
		if testCase.coil.falling() != testCase.expected {
			t.Fail()
		}
	}
}

func TestCoilUpdate(t *testing.T) {
	table := []struct {
		value            bool
		initialPrevious  bool
		initialCurrent   bool
		expectedPrevious bool
		expectedCurrent  bool
		shouldPublish    bool
		switchType       SwitchType
	}{
		// Update in case of true value NO
		{
			value:            true,
			initialPrevious:  false,
			initialCurrent:   false,
			expectedPrevious: false,
			expectedCurrent:  true,
			shouldPublish:    true,
			switchType:       NO,
		},
		// Update in case of a false value NC
		{
			value:            false,
			initialPrevious:  true,
			initialCurrent:   true,
			expectedPrevious: true,
			expectedCurrent:  false,
			shouldPublish:    true,
			switchType:       NC,
		},
		// No update in case of false value NO
		{
			value:            false,
			initialPrevious:  true,
			initialCurrent:   true,
			expectedPrevious: true,
			expectedCurrent:  false,
			shouldPublish:    false,
			switchType:       NO,
		},
		// No update in case of true value NC
		{
			value:            true,
			initialPrevious:  false,
			initialCurrent:   false,
			expectedPrevious: false,
			expectedCurrent:  true,
			shouldPublish:    false,
			switchType:       NC,
		},
	}
	for _, testCase := range table {
		// Set up the coil
		coil := Coil{previous: testCase.initialPrevious, current: testCase.initialCurrent, switchType: testCase.switchType}
		// Set up mqtt client (conditionally)
		mqttClient := &mocks.MQTTClient{}
		if testCase.shouldPublish {
			mqttClient.On("Publish", mock.AnythingOfType("string"), byte(0), false, "trigger").Return(&mqtt.PublishToken{})
		}
		// Do the actual call
		coil.Update(testCase.value, mqttClient)
		// Check: conditionally check for mqtt client called
		if testCase.shouldPublish {
			mqttClient.AssertExpectations(t)
		}
		// Check previous value
		if coil.previous != testCase.expectedPrevious {
			t.Errorf("Expected previous %v but got %v\n", testCase.expectedPrevious, coil.previous)
		}
		// Check current value
		if coil.current != testCase.expectedCurrent {
			t.Errorf("Expected current %v but got %v\n", testCase.expectedCurrent, coil.current)
		}
	}
}
