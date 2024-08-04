package main

import (
	"encoding/xml"
	"os"
)

func GetConfig() Config {
	data, err := os.ReadFile("config.xml")
	checkError(err)

	var config Config
	err = xml.Unmarshal(data, &config)
	checkError(err)

	return config
}
