package main

import "encoding/json"

func loadConfig (filename string, config * interface{}) error{
	jsonFileBody,err := ReadToString(filename)
	if err != nil {
		return err
	}
	err = updateStateFromJson(config,jsonFileBody)
	return err
}

func loadClientConfig (filename string) (* ClientNetworkConfig,error) {
	var config ClientNetworkConfig
	jsonFileBody,err := ReadToString(filename)
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal([]byte(jsonFileBody),&config)
	return &config, err
}

func loadServerConfig (filename string) (* ServerNetworkConfig,error) {
	var config ServerNetworkConfig
	jsonFileBody,err := ReadToString(filename)
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal([]byte(jsonFileBody),&config)
	return &config, err
}
