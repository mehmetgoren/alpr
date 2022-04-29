package utils

import (
	"encoding/json"
	"log"
)

func DeserializeJson(data string, result interface{}) error {
	err := json.Unmarshal([]byte(data), result)
	if err != nil {
		log.Println("Error deserializing json: ", err)
		return err
	}
	return nil
}
