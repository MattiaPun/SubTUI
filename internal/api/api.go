package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func generateSalt() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getAuthParams() url.Values {
	salt := generateSalt()
	hash := md5.Sum([]byte(AppConfig.Server.Password + salt))
	token := hex.EncodeToString(hash[:])

	v := url.Values{}
	v.Set("u", AppConfig.Server.Username)
	v.Set("t", token)
	v.Set("s", salt)
	v.Set("v", "1.16.1")
	v.Set("c", "SubTUI")
	v.Set("f", "json")

	return v
}

func subsonicGET(endpoint string, params map[string]string) (*SubsonicResponse, error) {
	baseUrl := AppConfig.Server.URL + "/rest" + endpoint

	v := getAuthParams()

	for key, value := range params {
		v.Set(key, value)
	}

	fullUrl := baseUrl + "?" + v.Encode()

	log.Printf("[API] Request: %s", fullUrl)
	resp, err := http.Get(fullUrl)
	if err != nil {
		log.Printf("[API] Connection Failed: %v", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[API] HTTP Error: %d | URL: %s", resp.StatusCode, fullUrl)
		return nil, fmt.Errorf("server error (HTTP %d)", resp.StatusCode)
	}

	var result SubsonicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func SubsonicLoginCheck() error {
	params := map[string]string{
		"username": AppConfig.Server.Username,
	}

	data, err := subsonicGET("/getUser", params)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}

	if data.Response.Status == "failed" && data.Response.Error != nil {
		if data.Response.Error.Code == 40 {
			return fmt.Errorf("wrong username or password")
		}
		return fmt.Errorf("api error: %s", data.Response.Error.Message)
	}

	if data.Response.User == nil && data.Response.Status == "ok" {
		return nil
	}

	return nil
}

func SubsonicSearchArtist(query string, page int) ([]Artist, error) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "150",
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
		"albumCount":   "150",
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
		"songCount":    "150",
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

func SubsonicGetAlbumList(searchType string) ([]Album, error) {
	params := map[string]string{
		"type": searchType,
		"size": "100",
	}

	data, err := subsonicGET("/getAlbumList", params)
	if err != nil {
		return nil, err
	}

	return data.Response.AlbumList.Albums, nil
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

	_, _ = subsonicGET("/star", params)
}

func SubsonicUnstar(id string) {
	params := map[string]string{
		"id": id,
	}

	_, _ = subsonicGET("/unstar", params)
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

func SubsonicRate(ID string, rating int) {
	params := map[string]string{
		"id":     ID,
		"rating": strconv.Itoa(rating),
	}

	_, _ = subsonicGET("/setRating", params)
}

func SubsonicStream(id string) string {
	baseUrl := AppConfig.Server.URL + "/rest/stream"

	v := getAuthParams()

	v.Set("id", id)
	v.Set("maxBitRate", "0")

	fullUrl := baseUrl + "?" + v.Encode()

	return fullUrl
}

func SubsonicScrobble(id string, submission bool) {
	time := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	params := map[string]string{
		"id":         id,
		"time":       time,
		"submission": strconv.FormatBool(submission),
	}

	_, _ = subsonicGET("/scrobble", params)
}

func SubsonicCoverArtUrl(id string, size int) string {
	baseUrl := AppConfig.Server.URL + "/rest/getCoverArt"

	v := getAuthParams()

	v.Set("id", id)
	v.Set("size", strconv.Itoa(size))

	url := baseUrl + "?" + v.Encode()
	return url
}

func SubsonicCoverArt(id string) ([]byte, error) {
	url := SubsonicCoverArtUrl(id, 50)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func SubsonicSaveQueue(ids []string, currentID string) {
	baseUrl := AppConfig.Server.URL + "/rest/savePlayQueue"

	v := getAuthParams()

	v.Set("current", currentID)
	for _, id := range ids {
		v.Add("id", id)
	}

	url := baseUrl + "?" + v.Encode()

	resp, _ := http.Get(url)
	defer func() { _ = resp.Body.Close() }()
}

func SubsonicGetQueue() (*PlayQueue, error) {
	params := map[string]string{}

	data, err := subsonicGET("/getPlayQueue", params)
	if err != nil {
		return nil, err
	}

	return &data.Response.PlayQueue, nil
}

func SubsonicAddToPlaylist(songID string, playlistID string) {
	params := map[string]string{
		"playlistId":  playlistID,
		"songIdToAdd": songID,
	}

	_, _ = subsonicGET("/updatePlaylist", params)
}

func SubsonicCreateShare(ID string) (string, error) {
	params := map[string]string{
		"id": ID,
	}

	data, err := subsonicGET("/createShare", params)
	if err != nil {
		log.Printf("[ERROR] API Error in CreateShare: %v", err)
		return "", err
	}

	url := data.Response.Shares.ShareList[0].URL
	log.Printf("[SHARE] Generated Share URL: %s", url)

	return url, nil
}
