package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func loadConfig(path string, out interface{}) {
	var err error
	data := []byte{}

	data, err = ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Couldn't read the file", err)
	}
	err = yaml.Unmarshal([]byte(data), out)
	if err != nil {
		log.Fatalf("loadConfig: %v", err)
	}
}
