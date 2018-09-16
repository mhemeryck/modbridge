package main

// ModbusMode indicates whether the involved register or coil is read-only or read and write allowed
type ModbusMode string

// Modbus mode constants
const (
	Read      ModbusMode = "R"  // Read-only
	ReadWrite            = "RW" // Both read and write allowed
	Write                = "W"  // Write-only. Added for own purposes
)

// RegisterDataType indicates how to read the contents of the register
type RegisterDataType string

// Register data type constants
const (
	MixedBits RegisterDataType = "mixedbits" // Register holds multiple contents
	Word                       = "word"      // Single word
)

// RegisterContent holds the description of the individual register bits
type RegisterContent struct {
	Slug   string
	Offset byte
	Length byte
}

// Register holds the description of the register part of a device modbus map
type Register struct {
	Address  uint16
	Mode     ModbusMode
	Datatype RegisterDataType
	Contents []RegisterContent
}

// Coil holds the description of the coil part of a device modbus map
type Coil struct {
	Address uint16
	Mode    ModbusMode
	Slug    string
}

// Configuration of modbridge
type Configuration struct {
	Registers       []Register
	Coils           []Coil
	MQTTBrokerURI   string `mapstructure:"mqtt_broker_uri"`
	MQTTClientID    string `mapstructure:"mqtt_client_id"`
	ModbusServerURI string `mapstructure:"modbus_server_uri"`
}
