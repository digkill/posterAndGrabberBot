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
	Video *Video `json:"video,omitempty"`
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

type Video struct {
	ID      int    `json:"id"`
	OwnerID int    `json:"owner_id"`
	Title   string `json:"title"`
	// ...
	Images []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"image"`
}
