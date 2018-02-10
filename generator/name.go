package generator

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Name struct {
	Raw        string
	Dash       string
	Snake      string
	Camel      string
	Lower      string
	LowerDash  string
	LowerSnake string
	LowerCamel string
	Title      string
	TitleSnake string
	TitleCamel string
	Upper      string
	UpperSnake string
	UpperCamel string
}

func NewName(in string) (name Name) {
	if strings.TrimSpace(in) == "" {
		return
	}

	camel := ToCamelCase(in)
	snake := ToSnakeCase(in)
	dash := ToDashCase(in)
	name = Name{
		Raw:        in,
		Dash:       dash,
		Camel:      camel,
		Snake:      snake,
		Lower:      strings.ToLower(in),
		LowerDash:  strings.ToLower(dash),
		LowerSnake: strings.ToLower(snake),
		LowerCamel: strings.ToLower(camel),
		Upper:      strings.ToUpper(in),
		UpperSnake: strings.ToUpper(snake),
		UpperCamel: strings.ToUpper(camel),
		Title:      strings.Title(in),
		TitleSnake: strings.Title(snake),
		TitleCamel: strings.Title(camel),
	}

	return
}

// Courtesy of https://github.com/etgryphon/stringUp/blob/master/stringUp.go
var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")

func ToCamelCase(src string) (out string) {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		}
	}
	out = string(bytes.Join(chunks, nil))
	out = strings.ToLower(string(out[0])) + string(out[1:])
	return out
}

// Courtesy of https://github.com/fatih/camelcase/blob/master/camelcase.go
func ToSnakeCase(src string) (out string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return src
	}
	entries := []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 && !strings.Contains(string(s), " ") {
			entries = append(entries, string(s))
		}
	}

	out = strings.ToLower(strings.Join(entries, "_"))

	for strings.Contains(out, "__") {
		out = strings.Replace(out, "__", "_", -1)
	}

	return out
}

func ToDashCase(in string) (out string) {
	return strings.Replace(ToSnakeCase(in), "_", "-", -1)
}
