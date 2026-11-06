package mobile

import (
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/SlavaNovikov/grif"
	"github.com/SlavaNovikov/grif/internal/pager"
	"github.com/SlavaNovikov/grif/migrations"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/server"
	"github.com/SlavaNovikov/grif/pkg/utils"
)

// StartServer запускает ваш HTTP-обработчик в фоне.
// Функция вызывается один раз при старте iOS-приложения.
func StartServer(port int, dbPath string) {
	go func() {

		// 1. Инициализируем БД (создаем таблицы, если их нет)
		// dbPath на iPhone должен быть путем к папке Documents (см. ниже)
		db := database.NewSQLite(dbPath, migrations.InitSQL)
		defer db.Close()

		tmpl, err := template.New("").Funcs(utils.FuncMap).ParseFS(grif.TemplatesFS, "templates/*.gohtml")
		if err != nil {
			log.Fatal(err)
		}

		mux := pager.New(db, tmpl, pager.Nav, pager.Registry).NewServeMux()
		handler := pager.StaticMiddleware(mux)
		server.Serve(server.Config{Url: fmt.Sprintf("127.0.0.1:%d", port)}, time.Now(), handler)
	}()
}
