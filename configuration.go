package modbridge

// ModbusMode indicates whether the involved register or coil is read-only or read and write allowed
type ModbusMode string

// Modbus mode constants
const (
	Read      ModbusMode = "R"  // Read-only
	ReadWrite            = "RW" // Both read and write allowed
	Write                = "W"  // Write-only. Added for own purposes
)

// CoilConfig holds the description of the coil part of a device modbus map
type CoilConfig struct {
	Address uint16
	Mode    ModbusMode
	Slug    string
}

// isWriteOnly indicates whether a given CoilConfig is write-only
func (coilConfig *CoilConfig) isWriteOnly() bool {
	return coilConfig.Mode == Write
}

// Configuration of modbridge
type Configuration struct {
	Coils           []CoilConfig
	MQTTBrokerURI   string `yaml:"mqtt_broker_uri"`
	MQTTClientID    string `yaml:"mqtt_client_id"`
	ModbusServerURI string `yaml:"modbus_server_uri"`
}

// filterCoilConfigs applies a filter based on a test function passed in
func (c *Configuration) filterCoilConfig() (filtered []CoilConfig) {
	for _, coilConfig := range c.Coils {
		if !coilConfig.isWriteOnly() {
			filtered = append(filtered, coilConfig)
		}
	}
	return
}

// CoilsList generates a list of non-write only coils from a configuration object
func (c *Configuration) CoilsList() (coils []Coil) {
	for _, coilConfig := range c.filterCoilConfig() {
		coils = append(coils, Coil{Address: coilConfig.Address, Slug: coilConfig.Slug, switchType: NO})
	}
	return
}

// CoilsMap generates a reverse mapping of the Slugs back to the original coil
func (c *Configuration) CoilsMap() (coils map[string]Coil) {
	coils = make(map[string]Coil)
	for _, coilConfig := range c.Coils {
		coils[coilConfig.Slug] = Coil{Address: coilConfig.Address, Slug: coilConfig.Slug, switchType: NO}
	}
	return
}

// CoilGroupsList generates a list of groups, out of the filtered list obtained from the config
func (c *Configuration) CoilGroupsList() []CoilGroup {
	return GroupCoils(c.CoilsList())
}
