package splitpt_client

import (
	"errors"
	"log"

	"github.com/BurntSushi/toml"
)

type SplitPTConfig struct {
	Connections  map[string]Connections
	SplittingAlg string
}

type Connections struct {
	Transport string
	Args      []string
	Cert      string
}

func GetClientTOMLConfig(tomlFilename string) (*ConnectionsList, error) {
	log.Printf("Decoding TOML")
	log.Printf(tomlFilename)
	var config ConnectionsList
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
		return _, errors.New("Invalid splitting algorithm in TOML")
	}
	log.Println(meta.Keys())
	log.Println(meta.Undecoded())
	return &config, nil
}
