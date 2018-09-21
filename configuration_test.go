package modbridge

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestConfiguration(t *testing.T) {
	input := []byte(`# Example config
coils:
- Address: 0
  mode: "W"
  slug: "digital-output-1-1"
- Address: 10
  mode: "R"
  slug: "digital-input-1-1"
mqtt_broker_uri: "tcp://mqtt:1883"
mqtt_client_id: "modbridge"
modbus_server_uri: "modbus:502"`)
	var c Configuration
	err := yaml.Unmarshal(input, &c)
	if err != nil {
		t.Errorf("Expected no errors parsing example config, got %v\n", err)
	}
	if len(c.Coils) != 2 {
		t.Errorf("Expected parsing 2 coils, got %d\n", len(c.Coils))
	}
	if c.MQTTBrokerURI != "tcp://mqtt:1883" {
		t.Errorf("Expected parsing MQTT broker URI, got %v\n", c.MQTTBrokerURI)
	}
	if c.MQTTClientID != "modbridge" {
		t.Errorf("Expected parsing MQTT client ID, got %v\n", c.MQTTClientID)
	}
	if c.ModbusServerURI != "modbus:502" {
		t.Errorf("Expected parsing modbus server URI, got %v\n", c.ModbusServerURI)
	}
}

func TestCoilsListConfiguration(t *testing.T) {
	c := Configuration{
		Coils: []CoilConfig{
			CoilConfig{Address: 1, Mode: Read}, CoilConfig{Address: 2, Mode: ReadWrite}, CoilConfig{Address: 3, Mode: Write},
		},
	}
	coils := c.CoilsList()
	if coils[0].Address != 1 || coils[1].Address != 2 || len(coils) != 2 {
		t.Errorf("Retrieving list of coils from config failed, got %v\n", coils)
	}
}

func TestCoilsMapConfiguration(t *testing.T) {
	c := Configuration{
		Coils: []CoilConfig{
			CoilConfig{Slug: "a", Mode: Read}, CoilConfig{Slug: "b", Mode: ReadWrite}, CoilConfig{Slug: "c", Mode: Write},
		},
	}
	coils := c.CoilsMap()
	if _, ok := coils["a"]; !ok {
		t.Errorf("Expected a coil mapped to a, not found\n")
	}
	if _, ok := coils["b"]; !ok {
		t.Errorf("Expected a coil mapped to b, not found\n")
	}
	if _, ok := coils["c"]; !ok {
		t.Errorf("Expected a coil mapped to c, not found\n")
	}
}

func TestCoilGroupsListConfiguration(t *testing.T) {
	c := Configuration{
		Coils: []CoilConfig{
			CoilConfig{Address: 1, Mode: Read}, CoilConfig{Address: 10, Mode: ReadWrite}, CoilConfig{Address: 20, Mode: Write},
		},
	}
	expected := []CoilGroup{
		CoilGroup{offset: 1, coils: []Coil{Coil{Address: 1}}},
		CoilGroup{offset: 10, coils: []Coil{Coil{Address: 10}}},
	}
	actual := c.CoilGroupsList()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v\n", expected, actual)
	}
}
