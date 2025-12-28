package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type SubsonicResponse struct {
	Response struct {
		Status            string        `json:"status"`
		SearchResult      SearchResult3 `json:"searchResult3"`
		PlaylistContainer struct {
			Playlists []Playlist `json:"playlist"`
		} `json:"playlists"`
		PlaylistDetail struct {
			Entries []Song `json:"entry"`
		} `json:"playlist"`
		Album struct {
			Songs []Song `json:"song"`
		} `json:"album"`
		Artist struct {
			Albums []Album `json:"album"`
		} `json:"artist"`
		Starred2 struct {
			Artist []Artist `json:"artist"`
			Album  []Album  `json:"album"`
			Song   []Song   `json:"song"`
		} `json:"starred2"`
	} `json:"subsonic-response"`
}

type SearchResult3 struct {
	Artists []Artist `json:"artist"`
	Albums  []Album  `json:"album"`
	Songs   []Song   `json:"song"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
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

func generateSalt() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func subsonicGET(endpoint string, params map[string]string) (*SubsonicResponse, error) {
	baseUrl := "https://" + AppConfig.Domain + "/rest" + endpoint

	salt := generateSalt()
	hash := md5.Sum([]byte(AppConfig.Password + salt))
	token := hex.EncodeToString(hash[:])

	v := url.Values{}
	v.Set("u", AppConfig.Username)
	v.Set("t", token)
	v.Set("s", salt)
	v.Set("v", "1.16.1")
	v.Set("c", "DepthTUI")
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

func SubsonicPing() error {
	data, err := subsonicGET("/ping", nil)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}

	if data.Response.Status != "ok" {
		return fmt.Errorf("authentication failed: server returned status %s", data.Response.Status)
	}

	return nil
}

func SubsonicSearchArtist(query string, page int) ([]Artist, error) {
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
		return nil, err
	}

	return data.Response.SearchResult.Artists, nil
}

func SubsonicSearchAlbum(query string, page int) ([]Album, error) {
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
		return nil, err
	}

	return data.Response.SearchResult.Albums, nil
}

func SubsonicSearchSong(query string, page int) ([]Song, error) {
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
		return nil, err
	}

	return data.Response.SearchResult.Songs, nil
}

func SubsonicGetPlaylistSongs(id string) ([]Song, error) {
	params := map[string]string{
		"id": id,
	}

	data, err := subsonicGET("/getPlaylist", params)
	if err != nil {
		return nil, err
	}

	return data.Response.PlaylistDetail.Entries, nil
}

func SubsonicGetPlaylists() ([]Playlist, error) {
	params := map[string]string{}

	data, err := subsonicGET("/getPlaylists", params)
	if err != nil {
		return nil, err
	}

	return data.Response.PlaylistContainer.Playlists, nil
}

func SubsonicGetAlbum(id string) ([]Song, error) {
	params := map[string]string{
		"id": id,
	}

	data, err := subsonicGET("/getAlbum", params)
	if err != nil {
		return nil, err
	}

	return data.Response.Album.Songs, nil
}

func SubsonicGetArtist(id string) ([]Album, error) {
	params := map[string]string{
		"id": id,
	}

	data, err := subsonicGET("/getArtist", params)
	if err != nil {
		return nil, err
	}

	return data.Response.Artist.Albums, nil
}

func SubsonicStar(id string) {
	params := map[string]string{
		"id": id,
	}

	subsonicGET("/star", params)
}

func SubsonicUnstar(id string) {
	params := map[string]string{
		"id": id,
	}

	subsonicGET("/unstar", params)
}

func SubsonicGetStarred() (*SearchResult3, error) {
	data, err := subsonicGET("/getStarred2", nil)
	if err != nil {
		return nil, err
	}

	return &SearchResult3{
		Artists: data.Response.Starred2.Artist,
		Albums:  data.Response.Starred2.Album,
		Songs:   data.Response.Starred2.Song,
	}, nil
}

func SubsonicStream(id string) string {
	baseUrl := "https://" + AppConfig.Domain + "/rest/stream"

	salt := generateSalt()
	hash := md5.Sum([]byte(AppConfig.Password + salt))
	token := hex.EncodeToString(hash[:])

	v := url.Values{}
	v.Set("id", id)
	v.Set("maxBitRate", "0")
	v.Set("u", AppConfig.Username)
	v.Set("t", token)
	v.Set("s", salt)
	v.Set("v", "1.16.1")
	v.Set("c", "DepthTUI")
	v.Set("f", "json")

	fullUrl := baseUrl + "?" + v.Encode()

	return fullUrl
}

func SubsonicScrobble(id string) {
	time := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	params := map[string]string{
		"id":         id,
		"time":       time,
		"submission": "0",
	}

	subsonicGET("/scrobble", params)
}
