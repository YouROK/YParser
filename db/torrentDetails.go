package db

import "time"

const (
	CatMovie         = "Movie"
	CatSeries        = "Series"
	CatDocMovie      = "DocMovie"
	CatDocSeries     = "DocSeries"
	CatCartoonMovie  = "CartoonMovie"
	CatCartoonSeries = "CartoonSeries"
	CatTVShow        = "TVShow"
	CatAnime         = "Anime"
)

type TorrentDetails struct {
	Title      string
	Name       string
	OrigName   string
	Categories []string
	Size       string
	CreateDate time.Time
	UpdateTime time.Time
	Tracker    string
	Link       string
	Year       int
	Peer       int
	Seed       int
	Magnet     string
}
