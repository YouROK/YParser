package parser

import (
	"YParser/bencode"
	"YParser/config"
	"YParser/db"
	"YParser/tasker"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var bitruHost = "https://bitru.org"

type BitruParser struct {
	*Parser
	mu           sync.Mutex
	magnetTasker tasker.Tasker
}

func NewBitru() *BitruParser {
	prs := new(Parser)
	br := new(BitruParser)
	br.Parser = prs
	br.Tasks.Threaded = false
	return br
}

func (self *BitruParser) Parse() {
	if self.BeginParse() {
		defer self.EndParse()

		pages := self.readCategories()
		self.addAll(pages)
		self.Tasks.Run()
	}
}

//читаем категории и заносим в таски что парсить
func (self *BitruParser) readCategories() map[string]int {
	log.Println("Read BitRu categories")

	var categories = []string{"movie", "serial"}
	var pages = map[string]int{}

	for _, cat := range categories {
		link := bitruHost + "/browse.php?tmp=" + cat
		log.Println("Read", link)
		body, err := get(link)
		if err == nil {
			re, err := regexp.Compile(`<a href="browse\.php\?tmp=` + cat + `&page=\d+#content">(\d+)<\/a>`)
			if err != nil {
				log.Fatalf("Error compile regex %v", err)
			}
			matches := re.FindAllStringSubmatch(body, -1)
			if len(matches) > 1 {
				pgs, err := strconv.Atoi(matches[len(matches)-1][1])
				if err == nil {
					self.mu.Lock()
					pages[cat] = pgs
					log.Println("Category readed", link, pgs)
					self.mu.Unlock()
				}
			}
		}
	}

	return pages
}

func (self *BitruParser) addAll(pages map[string]int) {
	var newTasks []*tasker.TaskParser
	for cat, pgs := range pages {
		for i := 1; i <= pgs; i++ {
			page := strconv.Itoa(i)
			link := bitruHost + "/browse.php?tmp=" + cat + "&page=" + page
			task := self.FindTask(link)
			if task == nil || time.Now().Before(task.UpdateTime.Add(time.Duration(config.Config.BitruParseTime)*time.Hour)) {
				// нет задачи или прошло больше X часов, добавляем в очередь
				newTasks = append(newTasks, &tasker.TaskParser{
					UpdateTime: time.Now(),
					Link:       link,
					Category:   cat,
					Worker:     self.ParsePage,
				})
			}
		}
	}
	self.Tasks.Tasks = newTasks
}

func (self *BitruParser) ParsePage(task *tasker.TaskParser) {
	// Парсим страницу с торрентами
	//time.Sleep(time.Second)
	body, err := get(task.Link)
	if err != nil {
		log.Println("Error get page:", err, task.Link)
		return
	}
	log.Println("Readed:", task.Link)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(body))
	if err != nil {
		log.Println("Error parse page:", err, task.Link)
		return
	}

	doc.Find("div#system").Find("tbody").Find("tr").Each(func(_ int, selection *goquery.Selection) {
		selTd := selection.Find("td")

		itm := new(db.TorrentDetails)
		itm.Title = selTd.Find("div.b-title").Find("a.main").Text()
		info := selTd.Find("div.b-info").Find("span").Text()
		arr := strings.Split(info, " от")
		if len(arr) > 0 {
			itm.CreateDate = self.parseDate(arr[0])
		} else {
			log.Println("Error parse date", info)
		}

		genres := selTd.Find("div.b-info").Find("a").Text()

		self.parseTitle(itm, task, genres)

		itm.Size = node2Text(selTd.Get(2))
		itm.Seed, _ = strconv.Atoi(strings.TrimSpace(node2Text(selTd.Get(3))))
		itm.Peer, _ = strconv.Atoi(strings.TrimSpace(node2Text(selTd.Get(4))))
		itm.Tracker = "Bitru"

		torrFilePath, _ := selTd.Find("div.b-title").Find("a.main").Attr("href")
		Url, _ := url.Parse(bitruHost + "/" + torrFilePath)
		id := Url.Query().Get("id")
		itm.Magnet = self.getMagnet("https://bitru.org/download.php?id="+id, bitruHost+"/"+torrFilePath)
		if itm.Magnet != "" {
			db.Add(itm)
		}
	})
	task.UpdateTime = time.Now()
}

func (self *BitruParser) getMagnet(link, referer string) string {
	buf, err := getBuf(link, referer)
	if err != nil {
		return ""
	}
	return bencode.ToMagnet(buf)
}

func (self *BitruParser) parseTitle(td *db.TorrentDetails, task *tasker.TaskParser, genres string) {
	re, err := regexp.Compile("\\d+? сезон ")
	if err != nil {
		log.Fatalf("Error parse torrent name:", err)
	}
	title := re.ReplaceAllString(td.Title, "")

	re, err = regexp.Compile("\\(1-\\d*? из .+?\\) ")
	if err != nil {
		log.Fatalf("Error parse torrent name:", err)
	}
	title = re.ReplaceAllString(title, "")

	re, err = regexp.Compile("(.+)\\((.+)\\)")
	if err != nil {
		log.Fatalf("Error parse torrent name:", err)
	}
	matches := re.FindStringSubmatch(title)
	if len(matches) > 2 {
		if strings.Contains(matches[2], "-") {
			arr := strings.Split(matches[2], "-")
			if len(arr) > 0 {
				matches[2] = strings.TrimSpace(arr[0])
			}
		}
		yr, _ := strconv.Atoi(strings.TrimSpace(matches[2]))
		td.Year = yr
		arr := strings.Split(matches[1], "/")
		if len(arr) == 1 {
			td.Name = arr[0]
		} else if len(arr) > 1 {
			td.Name = arr[0]
			td.OrigName = arr[len(arr)-1]
		}
		td.Name = strings.TrimSpace(td.Name)
		td.OrigName = strings.TrimSpace(td.OrigName)
	}

	if strings.Contains(genres, "Мультфильм") {
		if task.Category == "movie" {
			td.Categories = append(td.Categories, db.CatCartoonMovie)
		} else {
			td.Categories = append(td.Categories, db.CatCartoonSeries)
		}
	} else if strings.Contains(genres, "Аниме") {
		td.Categories = append(td.Categories, db.CatAnime)
	} else {
		if task.Category == "movie" {
			td.Categories = append(td.Categories, db.CatMovie)
		} else {
			td.Categories = append(td.Categories, db.CatSeries)
		}
	}
}

func (self *BitruParser) parseDate(date string) time.Time {
	if strings.Contains(date, "Сегодня") {
		dd := time.Now()
		arr := strings.Split(date, " ")
		if len(arr) != 3 {
			return time.Now()
		}
		timetmp := arr[2]
		arr = strings.Split(timetmp, ":")
		if len(arr) != 2 {
			return time.Now()
		}
		hh, _ := strconv.Atoi(arr[0])
		mm, _ := strconv.Atoi(arr[1])
		return time.Date(dd.Year(), dd.Month(), dd.Day(), hh, mm, 0, 0, time.Now().Location())
	}
	if strings.Contains(date, "Вчера") {
		dd := time.Now().AddDate(0, 0, -1)
		arr := strings.Split(date, " ")
		if len(arr) != 3 {
			return time.Now()
		}
		timetmp := arr[2]
		arr = strings.Split(timetmp, ":")
		if len(arr) != 2 {
			return time.Now()
		}
		hh, _ := strconv.Atoi(arr[0])
		mm, _ := strconv.Atoi(arr[1])
		return time.Date(dd.Year(), dd.Month(), dd.Day(), hh, mm, 0, 0, time.Now().Location())
	}

	var monthsMap = map[string]int{
		"января": 1, "февраля": 2, "марта": 3,
		"апреля": 4, "мая": 5, "июня": 6,
		"июля": 7, "августа": 8, "сентября": 9,
		"октября": 10, "ноября": 11, "декабря": 12,
	}

	darr := strings.Split(date, " ")
	if len(darr) != 5 {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())
	}

	day, _ := strconv.Atoi(darr[0])
	month, _ := monthsMap[darr[1]]
	year, _ := strconv.Atoi(darr[2])
	timetmp := darr[4]
	arr := strings.Split(timetmp, ":")
	if len(arr) != 2 {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
	}
	hh, _ := strconv.Atoi(arr[0])
	mm, _ := strconv.Atoi(arr[1])

	return time.Date(year, time.Month(month), day, hh, mm, 0, 0, time.Now().Location())
}
