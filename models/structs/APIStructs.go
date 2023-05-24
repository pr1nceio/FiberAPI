package structs

import "encoding/json"

type MusicResponse struct {
	Status string      `json:"status"`
	Name   string      `json:"name"`
	Artist string      `json:"artist"`
	Size   json.Number `json:"size"`
	Url    string      `json:"url"`
}

type PaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Url     string `json:"url"`
}
