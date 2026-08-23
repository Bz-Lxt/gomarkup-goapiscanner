package fingerprint

import (
	"sync"

	"github.com/alkaid/goapiscanner/internal/model"
)

// Matcher is a hand-written fingerprint engine. Rules are evaluated in order;
// the first hit for a (endpoint, class, param) triple wins to suppress duplicates.
type Matcher struct {
	mu    sync.Mutex
	seen  map[string]struct{}
	rules []Rule
}

func New() *Matcher {
	return &Matcher{
		seen:  make(map[string]struct{}),
		rules: defaultRules(),
	}
}

func (m *Matcher) Reset() {
	m.mu.Lock()
	m.seen = make(map[string]struct{})
	m.mu.Unlock()
}

func (m *Matcher) Match(p Probe) (Match, bool) {
	for _, r := range m.rules {
		hit := applyRule(r, p)
		if !hit.Hit {
			continue
		}
		key := p.Method + "|" + p.Endpoint + "|" + string(hit.Class) + "|" + p.ParamName
		m.mu.Lock()
		if _, ok := m.seen[key]; ok {
			m.mu.Unlock()
			return Match{}, false
		}
		m.seen[key] = struct{}{}
		m.mu.Unlock()
		return hit, true
	}
	return Match{}, false
}

func (m *Matcher) SeenCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.seen)
}

func FindingOf(taskID string, p Probe, hit Match, id, ts string) model.Finding {
	return model.Finding{
		ID:         id,
		TaskID:     taskID,
		Endpoint:   p.Endpoint,
		Method:     p.Method,
		Class:      hit.Class,
		Severity:   hit.Severity,
		Title:      hit.Title,
		Evidence:   hit.Evidence,
		Payload:    p.Payload,
		ParamName:  p.ParamName,
		StatusCode: p.StatusCode,
		LatencyMS:  p.Latency.Milliseconds(),
		Advice:     hit.Advice,
		CreatedAt:  ts,
	}
}
