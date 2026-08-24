package payload

import "github.com/alkaid/goapiscanner/internal/model"

type Item struct {
	Class   model.VulnClass
	Value   string
	Hint    string
	Timeout bool // timing payloads expect elevated client timeout
}

func Catalog() []Item {
	out := make([]Item, 0, 80)
	out = append(out, sqli()...)
	out = append(out, timeBlind()...)
	out = append(out, xss()...)
	out = append(out, unauth()...)
	out = append(out, traversal()...)
	out = append(out, cmdi()...)
	return out
}

func ByClass(c model.VulnClass) []Item {
	var out []Item
	for _, it := range Catalog() {
		if it.Class == c {
			out = append(out, it)
		}
	}
	return out
}
