package splitpt_client

import (
	"errors"
	"log"

	"github.com/BurntSushi/toml"
)

type SplitPTConfig struct {
	SplittingAlg string
	Connections  map[string][]struct {
		Transport string
		Args      []string
		Cert      string
	}
}

func GetClientTOMLConfig(tomlFilename string) (*SplitPTConfig, error) {
	log.Printf("Decoding TOML")
	log.Printf(tomlFilename)
	var config SplitPTConfig
	meta, err := toml.DecodeFile(tomlFilename, &config)
	if err != nil {
		log.Printf("Error decoding TOML config")
		return nil, err
	}

	switch config.SplittingAlg {
	case "round-robin":
		log.Printf(config.SplittingAlg)
	default:
		log.Printf("Invalid splitting algorithm")
		return nil, errors.New("Invalid splitting algorithm in TOML")
	}
	for conn := range config.Connections {
		log.Printf("Connections: ")
		log.Printf(conn)
	}
	log.Println(meta.Keys())
	log.Println(meta.Undecoded())
	return &config, nil
}
