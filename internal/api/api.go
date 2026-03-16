package api

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 20 * time.Second,
}

// Helper: Generate a random salt
func generateSalt() string {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(b)
}

// Helper: Compose the needed parameters
func getAuthParams() url.Values {
	v := url.Values{}
	v.Set("v", "1.16.1")
	v.Set("c", "SubTUI")
	v.Set("f", "json")

	switch strings.ToLower(AppServerConfig.Server.AuthMethod) {
	case "plaintext":
		salt := generateSalt()
		hash := md5.Sum([]byte(AppServerConfig.Server.Password + salt))
		token := hex.EncodeToString(hash[:])
		v.Set("u", AppServerConfig.Server.Username)
		v.Set("t", token)
		v.Set("s", salt)

	case "hashed":
		v.Set("u", AppServerConfig.Server.Username)
		v.Set("t", AppServerConfig.Server.PasswordToken)
		v.Set("s", AppServerConfig.Server.PasswordSalt)

	case "api_key":
		v.Set("apiKey", AppServerConfig.Server.ApiKey)
	}

	return v
}

// Helper: Redact sensitive parameters for debug log
func redactURL(rawUrl string) string {
	if !AppServerConfig.Security.RedactCredentialsInLogs {
		return rawUrl
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "<redacted_url>"
	}

	q := parsed.Query()
	if _, ok := q["t"]; ok {
		q.Set("t", "<REDACTED>")
	}
	if _, ok := q["s"]; ok {
		q.Set("s", "<REDACTED>")
	}
	if _, ok := q["p"]; ok {
		q.Set("p", "<REDACTED>")
	}
	if _, ok := q["apiKey"]; ok {
		q.Set("apiKey", "<REDACTED>")
	}

	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func subsonicGET(endpoint string, params map[string]string) (*SubsonicResponse, error) {
	baseUrl := AppServerConfig.Server.URL + "/rest" + endpoint

	v := getAuthParams()

	for key, value := range params {
		v.Set(key, value)
	}

	fullUrl := baseUrl + "?" + v.Encode()

	log.Printf("[API] Request: %s", redactURL(fullUrl))
	resp, err := httpClient.Get(fullUrl)
	if err != nil {
		log.Printf("[API] Connection Failed: %v", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[API] HTTP Error: %d | URL: %s", resp.StatusCode, redactURL(fullUrl))
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
		"username": AppServerConfig.Server.Username,
	}

	data, err := subsonicGET("/getUser", params)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}

	if data.Response.Status == "failed" && data.Response.Error != nil {
		if data.Response.Error.Code == 40 {
			return fmt.Errorf("invalid credentials")
		}
		return fmt.Errorf("api error: %s", data.Response.Error.Message)
	}

	if data.Response.User == nil && data.Response.Status == "ok" {
		return nil
	}

	return nil
}

func SubsonicSearchArtist(query string, offset int) ([]Artist, error) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "150",
		"artistOffset": strconv.Itoa(offset),
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

func SubsonicSearchAlbum(query string, offset int) ([]Album, error) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "0",
		"artistOffset": "0",
		"albumCount":   "150",
		"albumOffset":  strconv.Itoa(offset),
		"songCount":    "0",
		"songOffset":   "0",
	}

	data, err := subsonicGET("/search3", params)
	if err != nil {
		return nil, err
	}

	return data.Response.SearchResult.Albums, nil
}

func SubsonicSearchSong(query string, offset int) ([]Song, error) {
	params := map[string]string{
		"query":        query,
		"artistCount":  "0",
		"artistOffset": "0",
		"albumCount":   "0",
		"albumOffset":  "0",
		"songCount":    "150",
		"songOffset":   strconv.Itoa(offset),
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

func SubsonicGetAlbumList(searchType string, offset int) ([]Album, error) {
	params := map[string]string{
		"type":   searchType,
		"size":   "150",
		"offset": strconv.Itoa(offset),
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
	baseUrl := AppServerConfig.Server.URL + "/rest/stream"

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
	baseUrl := AppServerConfig.Server.URL + "/rest/getCoverArt"

	v := getAuthParams()

	v.Set("id", id)
	v.Set("size", strconv.Itoa(size))

	url := baseUrl + "?" + v.Encode()
	return url
}

func SubsonicCoverArt(id string, size int) ([]byte, error) {
	url := SubsonicCoverArtUrl(id, size)

	log.Printf("[API] Request: %s", redactURL(url))
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("[API] Failed to get cover Art: %v", err)
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
	baseUrl := AppServerConfig.Server.URL + "/rest/savePlayQueue"

	v := getAuthParams()

	v.Set("current", currentID)
	for _, id := range ids {
		v.Add("id", id)
	}

	url := baseUrl + "?" + v.Encode()

	log.Printf("[API] Request: %s", redactURL(url))
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("[API] Failed to save queue: %v", err)
		return
	}

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
