package generator

import (
	"strings"
	"text/template"
	"time"
)

var (
	funcMap = template.FuncMap{
		"TimeNowYear": time.Now().Year,
		"MinusP":      MinusP,
		"AddDB":       AddDB,
		"Add":         Add,
		"IsMap":       IsMap,
		"MinusStar":   MinusStar,
	}
)

func MinusP(in string) string {
	if string(in[len(in)-1]) == "P" {
		return in[:len(in)-1]
	}
	return in
}

func AddDB(in string) string {
	return in + "DB"
}

func Add(x, y int) int {
	return x + y
}

func IsMap(in string) bool {
	return strings.Contains(in, "map[")
}

func MinusStar(in string) string {
	return strings.Replace(in, "*", "", -1)
}
