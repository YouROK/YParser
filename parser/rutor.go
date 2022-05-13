package parser

import (
	"YParser/config"
	"YParser/db"
	"YParser/tasker"
	"YParser/utils"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var rutorHost = "http://rutor.lib"

type RutorParser struct {
	*Parser
	mu sync.Mutex
}

func NewRutor() *RutorParser {
	prs := new(Parser)
	rt := new(RutorParser)
	rt.Parser = prs
	rt.Tasks.Threaded = false
	return rt
}

func (self *RutorParser) Parse() {
	if self.BeginParse() {
		defer self.EndParse()

		pages := self.readCategories()
		self.addAll(pages)
		self.Tasks.Run()
	}
}

//читаем категории и заносим в таски что парсить
func (self *RutorParser) readCategories() map[string]int {
	// 1  - Зарубежные фильмы          | Фильмы
	// 5  - Наши фильмы                | Фильмы
	// 4  - Зарубежные сериалы         | Сериалы
	// 16 - Наши сериалы               | Сериалы
	// 12 - Научно-популярные фильмы   | Док. сериалы, Док. фильмы
	// 6  - Телевизор                  | ТВ Шоу
	// 7  - Мультипликация             | Мультфильмы, Мультсериалы
	// 10 - Аниме                      | Аниме
	// 17 - Иностранные релизы         | UA озвучка

	log.Println("Read Rutor categories")

	var categories = []string{"1", "5", "4", "16", "12", "6", "7", "10" /*"17"*/}
	var pages = map[string]int{}

	utils.PFor(categories, func(i int, cat string) {
		link := rutorHost + "/browse/0/" + cat + "/0/0"
		log.Println("Read", link)
		body, err := getNic(link)
		if err == nil {
			re, err := regexp.Compile("<a href=\"/browse/([0-9]+)/[0-9]+/[0-9]+/[0-9]+\"><b>[0-9]+&nbsp;-&nbsp;[0-9]+</b></a></p>")
			if err != nil {
				log.Fatalf("Error compile regex %v", err)
			}
			matches := re.FindStringSubmatch(body)
			if len(matches) > 1 {
				pgs, err := strconv.Atoi(matches[1])
				if err == nil {
					self.mu.Lock()
					pages[cat] = pgs
					log.Println("Category readed", link, pgs)
					self.mu.Unlock()
				}
			}
		}
	})

	return pages
}

func (self *RutorParser) addAll(pages map[string]int) {
	var newTasks []*tasker.TaskParser
	for cat, pgs := range pages {
		for i := 0; i < pgs; i++ {
			page := strconv.Itoa(i)
			link := rutorHost + "/browse/" + page + "/" + cat + "/0/0"
			task := self.FindTask(link)
			if task == nil || time.Now().Before(task.UpdateTime.Add(time.Duration(config.Config.RutorParseTime)*time.Hour)) {
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

func (self *RutorParser) ParsePage(task *tasker.TaskParser) {
	// Парсим страницу с торрентами
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

	doc.Find("div#index").Find("tr").Each(func(_ int, selection *goquery.Selection) {
		if selection.HasClass("backgr") {
			return
		}
		selTd := selection.Find("td")

		itm := new(db.TorrentDetails)
		itm.CreateDate = self.parseDate(node2Text(selTd.Get(0)))
		itm.Title = node2Text(selTd.Get(1))
		self.parseTitle(itm, task)
		itm.Magnet = selTd.Get(1).FirstChild.NextSibling.Attr[0].Val
		itm.Link = "http://rutor.info" + selTd.Get(1).LastChild.Attr[0].Val
		itm.UpdateTime = time.Now()
		if len(selTd.Nodes) == 4 {
			itm.Size = node2Text(selTd.Get(2))
			peers := node2Text(selTd.Get(3))
			prarr := strings.Split(peers, "  ")
			if len(prarr) > 1 {
				itm.Seed, _ = strconv.Atoi(prarr[0])
				itm.Peer, _ = strconv.Atoi(prarr[1])
			}
		} else if len(selTd.Nodes) == 5 {
			itm.Size = node2Text(selTd.Get(3))
			peers := node2Text(selTd.Get(4))
			prarr := strings.Split(peers, "  ")
			if len(prarr) > 1 {
				itm.Seed, _ = strconv.Atoi(prarr[0])
				itm.Peer, _ = strconv.Atoi(prarr[1])
			}
		}
		itm.Tracker = "Rutor"
		db.Add(itm)
	})
	task.UpdateTime = time.Now()
}

func (self *RutorParser) parseTitle(td *db.TorrentDetails, task *tasker.TaskParser) {
	re, err := regexp.Compile("(.+)\\((.+)\\)")
	if err != nil {
		log.Fatalf("Error parse torrent name:", err)
	}
	matches := re.FindStringSubmatch(td.Title)
	re, err = regexp.Compile("\\[.*?\\]")
	if len(matches) > 2 {
		yr, _ := strconv.Atoi(strings.TrimSpace(matches[2]))
		td.Year = yr
		arr := strings.Split(matches[1], "/")
		if len(arr) == 1 {
			td.Name = arr[0]
		} else if len(arr) > 1 {
			td.Name = arr[0]
			td.OrigName = arr[len(arr)-1]
		}
		if err == nil {
			td.Name = strings.TrimSpace(re.ReplaceAllString(td.Name, ""))
			td.OrigName = strings.TrimSpace(re.ReplaceAllString(td.OrigName, ""))
		}
	}
	title := td.Title
	if len(matches) > 0 {
		title = matches[1]
	}
	switch {
	case task.Category == "1", task.Category == "5":
		td.Categories = append(td.Categories, db.CatMovie)
	case task.Category == "4", task.Category == "16":
		td.Categories = append(td.Categories, db.CatSeries)
	case task.Category == "12":
		if re.MatchString(title) {
			td.Categories = append(td.Categories, db.CatDocSeries)
		} else {
			td.Categories = append(td.Categories, db.CatDocMovie)
		}
	case task.Category == "6":
		td.Categories = append(td.Categories, db.CatTVShow)
	case task.Category == "7":
		if re.MatchString(title) {
			td.Categories = append(td.Categories, db.CatCartoonSeries)
		} else {
			td.Categories = append(td.Categories, db.CatCartoonMovie)
		}
	case task.Category == "10":
		td.Categories = append(td.Categories, db.CatAnime)
	}
}

func (self *RutorParser) parseDate(date string) time.Time {
	var rutorMonth = map[string]int{
		"Янв": 1, "Фев": 2, "Мар": 3,
		"Апр": 4, "Май": 5, "Июн": 6,
		"Июл": 7, "Авг": 8, "Сен": 9,
		"Окт": 10, "Ноя": 11, "Дек": 12,
	}

	darr := strings.Split(date, " ")
	if len(darr) != 3 {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())
	}

	day, _ := strconv.Atoi(darr[0])
	month, _ := rutorMonth[darr[1]]
	year, _ := strconv.Atoi("20" + darr[2])

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
}
