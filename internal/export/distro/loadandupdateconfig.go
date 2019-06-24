/*
* Read and update the configuration

*/
package distro

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type configuration struct {
	Oedk oedk
	DataSource ds
}

type oedk struct {
	IsInitialized bool
	InitJson  string
}

type ds struct {
	Id string
	DataPoint []dp
}

type dp struct {
	Id string
}

// LoadConfigFromFile use to load toml configuration
func LoadConfigFromFile() (*configuration, error) {
	config := new(configuration)
	confDir := "./res/oedkconfig.toml"

	// Read toml file
	file, err := ioutil.ReadFile(confDir)
	if err != nil {
		return config, fmt.Errorf("could not load configuration file (%s): %v", confDir, err.Error())
	}

	// reformat toml to config struct as defined previously
	err = toml.Unmarshal(file, config)
	if err != nil {
		return config, fmt.Errorf("unable to parse configuration file (%s): %v", confDir, err.Error())
	}
	return config, err
}

// UpdateConfigFromFile use to store toml configuration
func UpdateConfigFromFile(config *configuration) error {
	confDir := "./res/oedkconfig.toml"

	// reformat config to bytes buffer
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return fmt.Errorf("unable to parse to configuration file (%s): %v", err.Error())
	}

	if err := ioutil.WriteFile(confDir, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("could not store configuration file (%s): %v", confDir, err.Error())
	}
	return  nil
}