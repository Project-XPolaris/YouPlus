package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func POSTRequestWithJSON(url string, data interface{}) (*http.Response, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(rawData))
	if err != nil {
		return nil, err
	}
	return response, nil
}
func UpdateRequestWithJSON(url string, data interface{}) (*http.Response, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(rawData))
	if err != nil {
		return nil, err
	}
	return response, nil
}
