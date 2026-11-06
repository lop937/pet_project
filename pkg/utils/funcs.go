package utils

import (
	"fmt"
	"html/template"
	"time"
)

// FuncMap — функции, доступные в шаблонах.
var FuncMap = template.FuncMap{
	"add": func(a, b int64) int64 { return a + b },
	"sub": func(a, b int64) int64 { return a - b },
	"mul": func(a, b int64) int64 { return a * b },
	"div": func(a, b int64) int64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"formatDate": func(layout string, t time.Time) string {
		return t.Format(layout)
	},
	// Args собирает map[string]interface{} из пар ключ-значение.
	// Пример: {{template "button_add" (Args "page" .route "id" 0)}}
	"Args": func(values ...interface{}) map[string]interface{} {
		m := make(map[string]interface{}, len(values)/2)
		for i := 0; i+1 < len(values); i += 2 {
			key := fmt.Sprint(values[i])
			m[key] = values[i+1]
		}
		return m
	},
}
