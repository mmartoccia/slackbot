package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Attributes(m interface{}) map[string]reflect.Type {
	typ := reflect.TypeOf(m)
	// if a pointer to a struct is passed, get the type of the dereferenced object
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// create an attribute data structure as a map of types keyed by a string.
	attrs := make(map[string]reflect.Type)
	// Only structs are supported so return an empty result if the passed object
	// isn't a struct
	if typ.Kind() != reflect.Struct {
		fmt.Printf("%v type can't have attributes inspected\n", typ.Kind())
		return attrs
	}

	// loop through the struct's fields and set the map
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			attrs[p.Name] = p.Type
		}
	}

	return attrs
}

func Underscore(s string) string {
	reg := regexp.MustCompile("([A-Z]+)([A-Z][a-z])")
	r := reg.ReplaceAllString(s, "${1}_${2}")

	reg = regexp.MustCompile("([a-z\\d])([A-Z])")
	r = reg.ReplaceAllString(r, "${1}_${2}")

	r = strings.ToLower(strings.Replace(r, "-", "_", -1))
	return r
}

func IsNumber(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}

	return false
}

func FormatHour(h int64) string {
	if h == 0 {
		return ""
	}

	v := float64(h) / 60
	return fmt.Sprintf("%.2f", v)
}
