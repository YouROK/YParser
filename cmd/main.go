package main

import (
	"YParser/config"
	"YParser/db"
	"YParser/parser"
	"YParser/proxy"
	"YParser/sheduler"
	"YParser/web"
)

func main() {
	//first load config
	config.Init()

	db.Init()
	proxy.InitProxy()

	rutorParser := parser.NewRutor()
	bitruParser := parser.NewBitru()

	sheduler.NewSheduler(5, func() {
		rutorParser.Parse()
	}).Start()

	sheduler.NewSheduler(5, func() {
		bitruParser.Parse()
	}).Start()

	web.WebInit()
}
