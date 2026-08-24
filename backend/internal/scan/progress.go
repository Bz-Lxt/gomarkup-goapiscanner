package scan

import "sync"

type Progress struct {
	mu    sync.Mutex
	Total int
	Sent  int
	Hits  int
	Crit  int
	High  int
	Med   int
	Low   int
	Info  int
}

func (p *Progress) AddSent() (sent, total, hits int) {
	p.mu.Lock()
	p.Sent++
	sent, total, hits = p.Sent, p.Total, p.Hits
	p.mu.Unlock()
	return
}

func (p *Progress) AddHit(rank int) (sent, total, hits int) {
	p.mu.Lock()
	p.Hits++
	switch rank {
	case 5:
		p.Crit++
	case 4:
		p.High++
	case 3:
		p.Med++
	case 2:
		p.Low++
	default:
		p.Info++
	}
	sent, total, hits = p.Sent, p.Total, p.Hits
	p.mu.Unlock()
	return
}

func (p *Progress) Snapshot() Progress {
	p.mu.Lock()
	defer p.mu.Unlock()
	return Progress{
		Total: p.Total,
		Sent:  p.Sent,
		Hits:  p.Hits,
		Crit:  p.Crit,
		High:  p.High,
		Med:   p.Med,
		Low:   p.Low,
		Info:  p.Info,
	}
}
