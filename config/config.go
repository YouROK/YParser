package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// db test redis 185.246.155.239:6379 redisdb@114328

type config struct {
	DBHost string `json:"db_host,omitempty"`
	DBPass string `json:"db_pass,omitempty"`

	WebPort string `json:"web_port,omitempty"`
	WebAddr string `json:"web_addr,omitempty"`

	RutorParseTime int `json:"rutor_parse_time"`
	BitruParseTime int `json:"bitru_parse_time"`

	HttpProxyList []string `json:"http_proxy_list,omitempty"`
}

var (
	Config *config
)

func Init() {
	dir := filepath.Dir(os.Args[0])
	file := filepath.Join(dir, "config.json")
	buf, err := ioutil.ReadFile(file)
	if err == nil {
		json.Unmarshal(buf, &Config)
	}
	if Config == nil {
		Config = new(config)
		//TODO change to localhost
		Config.DBHost = "185.246.155.239:6379"
		Config.DBPass = "redisdb@114328"

		Config.WebPort = "80"
		Config.WebAddr = ""

		Config.RutorParseTime = 5
		Config.BitruParseTime = 5

		buf, err := json.MarshalIndent(Config, "", " ")
		if err == nil {
			ioutil.WriteFile(file, buf, 0666)
		}
	}
}
