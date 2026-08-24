package payload

import "github.com/alkaid/goapiscanner/internal/model"

func unauth() []Item {
	return []Item{
		{Class: model.ClassUnauth, Value: "", Hint: "missing-auth"},
		{Class: model.ClassUnauth, Value: "Bearer invalid", Hint: "forged-bearer"},
		{Class: model.ClassUnauth, Value: "Bearer ", Hint: "empty-bearer"},
	}
}
