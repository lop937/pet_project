package utils

import (
	"net/url"
	"reflect"
	"strings"
)

// ExtractFields возвращает список имён полей структуры по тегу tag,
// объединённых запятой. Поля без тега и с тегом "-" пропускаются.
// Поддерживает указатель на структуру.
func ExtractFields(tag string, data interface{}) string {
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	t := v.Type()

	out := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tagVal := f.Tag.Get(tag)
		if tagVal == "" || tagVal == "-" {
			continue
		}
		// "name,omitempty" → "name"
		if idx := strings.Index(tagVal, ","); idx >= 0 {
			tagVal = tagVal[:idx]
		}
		if tagVal == "" {
			continue
		}
		out = append(out, tagVal)
	}
	return strings.Join(out, ",")
}

// GetFormFields оставляет из fields только те, что присутствуют в form,
// и возвращает их через запятую.
func GetFormFields(form url.Values, fields string) string {
	parts := strings.Split(fields, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := form[p]; ok {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}
