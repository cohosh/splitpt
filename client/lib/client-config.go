package splitpt_client

import (
	"log"

	"github.com/BurntSushi/toml"
)

type ConnectionsList struct {
	Connections  map[string]Connections
	Splittingalg string
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
	log.Printf(config.Splittingalg)
	log.Println(meta.Keys())
	log.Println(meta.Undecoded())
	return &config, nil
}
