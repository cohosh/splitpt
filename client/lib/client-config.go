package splitpt_client

import (
	"log"

	"github.com/BurntSushi/toml"
)

type ClientConfig struct {
	NumConnections       int
	ConnectionsTransport []string
}

func getClientConfig() (*ClientConfig, error) {
	var config ClientConfig
	_, err := toml.Decode("../splitpt-config.toml", &config)
	if err != nil {
		log.Printf("Error decoding TOML config")
		return nil, err
	}

	// Check that everything's in order
	if config.NumConnections != len(config.ConnectionsTransport) {
		log.Printf("TOML Config error: Number of connections does not match number of listed transports")
		return nil, err
	}

	return &config, nil
}
