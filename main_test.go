package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSlug(t *testing.T) {
	input := DigitalInput{description: "name with spaces"}
	if input.Slug() != "name_with_spaces" {
		t.Fail()
	}
}

func TestReadConfig(t *testing.T) {
	readConfig()
	keys := [3]string{"mqtt_broker_uri", "mqtt_client_id", "modbus_server_uri"}
	for _, key := range keys {
		if viper.GetString(key) == "" {
			t.Fail()
		}
	}
}

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
	modbusClient := Client{}
}
