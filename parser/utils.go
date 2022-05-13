package parser

import (
	"YParser/client"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"strings"
	"time"
)

func getNic(link string) (string, error) {
	var body string
	var err error
	for i := 0; i < 10; i++ {
		body, err = client.GetNic(link, "", "")
		if err == nil {
			break
		}
		log.Println("Error get page,tryes:", i+1, link)
		time.Sleep(time.Second * 2)
	}
	return body, err
}

func get(link string) (string, error) {
	var body string
	var err error
	for i := 0; i < 10; i++ {
		body, err = client.Get(link, "", "")
		if err == nil {
			break
		}
		log.Println("Error get page,tryes:", i+1, link)
		time.Sleep(time.Second * 2)
	}
	return body, err
}

func getBuf(link, referer string) ([]byte, error) {
	var body []byte
	var err error
	for i := 0; i < 10; i++ {
		body, err = client.GetBuf(link, referer, "")
		if err == nil {
			break
		}
		log.Println("Error get page,tryes:", i+1, link)
		time.Sleep(time.Second * 2)
	}
	return body, err
}

func node2Text(node *html.Node) string {
	return strings.TrimSpace(strings.Replace((&goquery.Selection{Nodes: []*html.Node{node}}).Text(), "\u00A0", " ", -1))
}

func replaceBadName(name string) string {
	name = strings.ReplaceAll(name, "Ванда/Вижн ", "ВандаВижн ")
	name = strings.ReplaceAll(name, "Ё", "Е")
	name = strings.ReplaceAll(name, "ё", "е")
	name = strings.ReplaceAll(name, "щ", "ш")
	return name
}
