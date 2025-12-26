package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const USERNAME = ""
const PASSWORD = ""
const SUBSONIC_DOMAIN = ""

type SubsonicResponse struct {
	Response struct {
		Status       string `json:"status"`
		SearchResult struct {
			Artists []Artist `json:"artist"`
			Albums  []Album  `json:"album"`
			Songs   []Song   `json:"song"`
		} `json:"searchResult3"`
	} `json:"subsonic-response"`
}

type SearchResult3 struct {
	Artists   []Artist   `json:"artist"`
	Albums    []Album    `json:"album"`
	Songs     []Song     `json:"song"`
	Playlists []Playlist `json:"playlist"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func subsonicGET(endpoint string, params map[string]string) (*SubsonicResponse, error) {
	baseUrl := "https://" + SUBSONIC_DOMAIN + "/rest" + endpoint

	v := url.Values{}
	v.Set("u", USERNAME)
	v.Set("p", PASSWORD)
	v.Set("v", "1.16.1")
	v.Set("c", "depth")
	v.Set("f", "json")

	for key, value := range params {
		v.Set(key, value)
	}

	fullUrl := baseUrl + "?" + v.Encode()

	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SubsonicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func subsonicSearchArtist(query string, page int) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "20",
		"artistOffset": strconv.Itoa(page * 20),
		"albumCount":   "0",
		"albumOffset":  "0",
		"songCount":    "0",
		"songOffset":   "0",
	}

	data, err := subsonicGET("/search3", params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, artist := range data.Response.SearchResult.Artists {
		fmt.Printf("Found Artist: %s\n", artist.Name)
	}
}

func subsonicSearchAlbum(query string, page int) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "0",
		"artistOffset": "0",
		"albumCount":   "20",
		"albumOffset":  strconv.Itoa(page * 20),
		"songCount":    "0",
		"songOffset":   "0",
	}

	data, err := subsonicGET("/search3", params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, album := range data.Response.SearchResult.Albums {
		fmt.Printf("Found Album: %s\n", album.Title)
	}
}

func subsonicSearchSong(query string, page int) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "0",
		"artistOffset": "0",
		"albumCount":   "0",
		"albumOffset":  "0",
		"songCount":    "20",
		"songOffset":   strconv.Itoa(page * 20),
	}

	data, err := subsonicGET("/search3", params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, song := range data.Response.SearchResult.Songs {
		fmt.Printf("Found Songs: %s\n", song.Title)
	}
}
