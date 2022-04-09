package main

func loadConfig (filename string, config interface{}) error{
	jsonFileBody,err := ReadToString(filename)
	if err != nil {
		return err
	}
	err = updateStateFromJson(config,jsonFileBody)
	return err
}
