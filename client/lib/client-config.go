package splitpt_client

import (
	"log"

	"github.com/BurntSushi/toml"
)

type PTConnection struct {
	Transport string
	Args      []string
	Cert      string
}

type ClientTOMLConfig struct {
	Connections  map[string]PTConnection
	SplittingAlg string
}

func GetClientTOMLConfig(tomlFilename string) (*ClientTOMLConfig, error) {
	var config ClientTOMLConfig
	_, err := toml.Decode(tomlFilename, &config)
	if err != nil {
		log.Printf("Error decoding TOML config")
		return nil, err
	}

	return &config, nil
}
