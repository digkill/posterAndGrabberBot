package models

import "time"

type WallGetResponse struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string

	Response struct {
		Items []struct {
			ID          int          `json:"id"`
			Attachments []Attachment `json:"attachments"`
		} `json:"items"`
	} `json:"response"`
}

type Attachment struct {
	Type  string `json:"type"`
	Photo *Photo `json:"photo,omitempty"`
}

type Photo struct {
	ID    int `json:"id"`
	Sizes []struct {
		Type   string `json:"type"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"sizes"`
}
