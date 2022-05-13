package proxy

import (
	"YParser/config"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ProxyList []string

type queryResp struct {
	Addr string
	Time time.Duration
}

func InitProxy() {
	if len(config.Config.HttpProxyList) == 0 {
		return
	}
	respChan := make(chan queryResp, 10)

	go func() {
		var wa sync.WaitGroup
		for _, addr := range config.Config.HttpProxyList {
			wa.Add(1)
			go func(a string) {
				query(a, respChan)
				wa.Done()
			}(addr)
		}
		wa.Wait()
		close(respChan)
	}()

	for r := range respChan {
		log.Println(r.Addr, r.Time.Seconds())
		ProxyList = append(ProxyList, r.Addr)
	}
}

func query(host string, c chan queryResp) {
	urlProxy := &url.URL{Host: host}
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(urlProxy)},
		Timeout:   120 * time.Second,
	}
	log.Println("Check proxy:", host)
	startTs := time.Now()
	resp, err := client.Get("http://rutor.info")
	if err != nil {
		log.Println("Check error:", host, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("Check error:", host, resp.StatusCode, resp.Status)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	timeDiff := time.Now().Sub(startTs)
	if strings.Contains(string(body), "<title>rutor.info :: Свободный торрент трекер</title>") {
		c <- queryResp{Addr: host, Time: timeDiff}
	}
}
