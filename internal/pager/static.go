package pager

import (
	"io/fs"
	"net/http"

	"github.com/SlavaNovikov/grif"
)

// StaticMiddleware отдаёт файлы из /static/ до того, как запрос попадёт в ServeMux.
func StaticMiddleware(next http.Handler) http.Handler {
	staticFS, _ := fs.Sub(grif.StaticFS, "static")
	fileServer := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/static/" {
			fileServer.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
