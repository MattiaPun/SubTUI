package api

type SubsonicResponse struct {
	Response struct {
		Status            string         `json:"status"`
		User              *SubsonicUser  `json:"user,omitempty"`
		Error             *SubsonicError `json:"error,omitempty"`
		SearchResult      SearchResult3  `json:"searchResult3"`
		PlaylistContainer struct {
			Playlists []Playlist `json:"playlist"`
		} `json:"playlists"`
		PlaylistDetail struct {
			Entries []Song `json:"entry"`
		} `json:"playlist"`
		Album struct {
			Songs []Song `json:"song"`
		} `json:"album"`
		AlbumList struct {
			Albums []Album `json:"album"`
		} `json:"albumList"`
		Artist struct {
			Albums []Album `json:"album"`
		} `json:"artist"`
		Starred2 struct {
			Artist []Artist `json:"artist"`
			Album  []Album  `json:"album"`
			Song   []Song   `json:"song"`
		} `json:"starred2"`
		PlayQueue PlayQueue `json:"playQueue"`
		Shares    struct {
			ShareList []struct {
				URL string `json:"url"`
			} `json:"share"`
		} `json:"shares"`
	} `json:"subsonic-response"`
}

type SubsonicUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SubsonicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PlayQueue struct {
	Current string `json:"current"`
	Entries []Song `json:"entry"`
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
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Duration int64  `json:"duration"`
}

type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	ArtistID string `json:"artistId"`
	Album    string `json:"album"`
	AlbumID  string `json:"albumId"`
	Duration int    `json:"duration"`
	Rating   int    `json:"userRating"`
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
