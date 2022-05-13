package db

import (
	"YParser/config"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
)

var (
	//torrents map[string]*TorrentDetails
	thread chan *TorrentDetails
	rdb    *redis.Client
	ctx    = context.Background()
)

func Init() {

	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Config.DBHost,
		Password: config.Config.DBPass,
		DB:       0,
	})

	//Read all torrents from db
	//log.Println("Load torrents from bd")
	//torrents = map[string]*TorrentDetails{}
	//iter := rdb.Scan(ctx, 0, "*", 100000).Iterator()
	//
	//for iter.Next(ctx) {
	//	key := iter.Val()
	//	val := rdb.Get(ctx, key)
	//	buf := val.Val()
	//	var torrent *TorrentDetails
	//	err := json.Unmarshal([]byte(buf), &torrent)
	//	if err == nil {
	//		torrents[torrent.Link] = torrent
	//	}
	//}
	//
	//log.Println("Loaded")
	// run worker
	worker()
}

func worker() {
	thread = make(chan *TorrentDetails, 1000)
	go func() {
		for true {
			t := <-thread
			buf, err := json.Marshal(t)
			if err != nil {
				panic(err)
			}
			hash := hex.EncodeToString(md5.New().Sum([]byte(t.Link)))
			key := strings.ToLower(t.Name+":"+t.OrigName+":"+strconv.Itoa(t.Year)) + ":" + hash
			err = rdb.Set(ctx, key, string(buf), 0).Err()
			if err != nil {
				panic(err)
			}
		}
	}()
}

func Search(query string) []*TorrentDetails {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	iter := rdb.Scan(ctx, 0, query, 10000).Iterator()
	var torrents []*TorrentDetails

	for iter.Next(ctx) {
		key := iter.Val()
		val := rdb.Get(ctx, key)
		buf := val.Val()
		var torrent *TorrentDetails
		err := json.Unmarshal([]byte(buf), &torrent)
		if err == nil {
			torrents = append(torrents, torrent)
		}
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}
	return torrents
}

func Add(t *TorrentDetails) {
	thread <- t
}
