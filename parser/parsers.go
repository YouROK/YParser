package parser

import (
	"YParser/tasker"
	"sync"
)

type Parser struct {
	Parse   func()
	Tasks   tasker.Tasker
	IsParse bool
	MU      sync.Mutex
}

func (p *Parser) BeginParse() bool {
	p.MU.Lock()
	if p.IsParse {
		return false
	}
	p.IsParse = true
	p.MU.Unlock()
	return true
}

func (p *Parser) EndParse() {
	p.IsParse = false
}

func (p *Parser) FindTask(link string) *tasker.TaskParser {
	for _, task := range p.Tasks.Tasks {
		if task.Link == link {
			return task
		}
	}
	return nil
}
