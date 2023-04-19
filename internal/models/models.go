package models

type BasicUserInfo struct {
	Id int64 `json:"id"`
}

type BasicHistoryInfo struct {
	Id           int64           `json:"id"`
	Title        string          `json:"title"`
	TitleShort   string          `json:"title_short"`
	TitleVersion string          `json:"title_version"`
	Link         string          `json:"link"`
	Duration     int64           `json:"duration"`
	Rank         int64           `json:"rank"`
	Artist       BasicArtistInfo `json:"artist"`
	Album        BasicAlbumInfo  `json:"album"`
	BasicImage   string
}

type BasicWrapHistory struct {
	Data []BasicHistoryInfo `json:"data"`
}

type BasicArtistInfo struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	NbAlbum       int64  `json:"nb_album"`
	NbFan         int64  `json:"nb_fan"`
}

type BasicAlbumInfo struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Cover       string `json:"cover"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
	CoverXl     string `json:"cover_xl"`
	Duration    int64  `json:"duration"`
	Fans        int64  `json:"fans"`
	ReleaseDate string `json:"release_date"`
}
