package main

import (
	"html/template"
	"log"
	"time"

	"github.com/SlavaNovikov/grif"
	"github.com/SlavaNovikov/grif/migrations"

	_ "github.com/mattn/go-sqlite3"

	"github.com/SlavaNovikov/grif/internal/pager"
	"github.com/SlavaNovikov/grif/pkg/database"

	"github.com/SlavaNovikov/grif/pkg/server"
	"github.com/SlavaNovikov/grif/pkg/utils"
)

func main() {
	tn := time.Now()
	// Инициализируем БД (создаем таблицы, если их нет)
	// dbPath на iPhone должен быть путем к папке Documents
	// в InitSQL код создания таблиц и миграции добавленный через embed
	db := database.NewSQLite("store.db", migrations.InitSQL)
	defer db.Close()

	// Важно: путь должен быть таким же, как в директиве //go:embed
	tmpl, err := template.New("").Funcs(utils.FuncMap).ParseFS(grif.TemplatesFS, "templates/*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	mux := pager.New(db, tmpl, pager.Nav, pager.Registry).NewServeMux()

	// Оборачиваем в middleware для статики
	handler := pager.StaticMiddleware(mux)

	server.Serve(server.Config{Url: "127.0.0.1:8081"}, tn, handler)
}
