package textproto

import (
	"sync"
)

type Pipeline struct {
	mu       sync.Mutex
	id       uint
	request  sequencer
	response sequencer
}

func (p *Pipeline) Next() uint {
	p.mu.Lock()
	id := p.id
	p.id++
	p.mu.Unlock()
	return id
}

func (p *Pipeline) StartRequest(id uint) {
	p.request.Start(id)
}

func (p *Pipeline) EndRequest(id uint) {
	p.request.End(id)
}

func (p *Pipeline) StartResponse(id uint) {
	p.response.Start(id)
}

func (p *Pipeline) EndReponse(id uint) {
	p.response.End(id)
}

type sequencer struct {
	mu   sync.Mutex
	id   uint
	wait map[uint]chan uint
}

func (s *sequencer) Start(id uint) {
	s.mu.Lock()
	if s.id == id {
		s.mu.Unlock()
		return
	}

	c := make(chan uint)
	if s.wait == nil {
		s.wait = make(map[uint]chan uint)
	}
	s.wait[id] = c
	s.mu.Unlock()
	<-c
}

func (s *sequencer) End(id uint) {
	s.mu.Lock()
	if s.id != id {
		panic("out of sync")
	}
	id++
	s.id = id
	if s.wait == nil {
		s.wait = make(map[uint]chan uint)
	}
	c, ok := s.wait[id]
	if ok {
		delete(s.wait, id)
	}
	s.mu.Unlock()
	if ok {
		c <- 1
	}
}
