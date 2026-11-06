package pager

import (
	"fmt"
	"html/template"
	"strings"

	"log"
	"net/http"
	"strconv"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/internal/export"
)

func New(store ds.Storage, tmpl *template.Template, nav []ds.RouteType, registry ds.Registry) Handler {
	return Handler{
		tmpl:     tmpl,
		store:    store,
		nav:      nav,
		registry: registry,
	}
}

type Handler struct {
	tmpl        *template.Template
	store       ds.Storage
	pageFactory Factory
	nav         []ds.RouteType
	registry    ds.Registry
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer func() {
		//	if rec := recover(); rec != nil {
		//		log.Println("FATAL RECOVERED panicMdl", rec, r.URL)
		//	}
		//}()
		next.ServeHTTP(w, r)
	})
}

func (h Handler) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Динамические страницы
	mux.Handle("/{page}", Recover(http.HandlerFunc(h.Content)))
	mux.Handle("/{page}/{id}", Recover(http.HandlerFunc(h.Form)))

	// Экспорт
	mux.HandleFunc("/export/works", func(w http.ResponseWriter, r *http.Request) {
		export.ExportWorks(w, r, h.store)
	})

	// Экспорт одной работы по ID
	mux.HandleFunc("/export/works/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/export/works/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Неверный ID работы", http.StatusBadRequest)
			return
		}
		export.ExportSingleWork(w, r, h.store, ds.ID(id))
	})

	// Экспорт одной работы
	//mux.HandleFunc("/export/works/", func(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из конца URL (например, /export/works/15)
	//	idStr := strings.TrimPrefix(r.URL.Path, "/export/works/")
	//	id, err := strconv.ParseInt(idStr, 10, 64)
	//	if err != nil {		http.Error(w, "Неверный ID работы", http.StatusBadRequest)
	//		return	}
	//	export.ExportSingleWork(w, r, h.store, ds.ID(id))
	//})

	mux.HandleFunc("/export/project/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := ds.IDFromReq(r)
		export.ExportProject(w, r, h.store, id)
	})
	mux.HandleFunc("/export/systems", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.ParseUint(r.URL.Query().Get("project_id"), 10, 64)
		export.ExportSystems(w, r, h.store, ds.ID(pid))
	})
	mux.HandleFunc("/export/full_project/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := ds.IDFromReq(r)
		export.ExportFullProject(w, r, h.store, id)
	})
	mux.HandleFunc("/export/work/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := ds.IDFromReq(r)
		export.ExportSingleWork(w, r, h.store, id)
	})

	mux.HandleFunc("/export/smeta/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := ds.IDFromReq(r)
		export.ExportFormattedProject(w, r, h.store, id)
	})

	mux.HandleFunc("/export/smeta-selected", func(w http.ResponseWriter, r *http.Request) {
		export.ExportSelectedWorks(w, r, h.store)
	})

	return mux
}

func (h Handler) Content(w http.ResponseWriter, r *http.Request) {
	page, obj, ok := h.checkPageAccess(w, r)
	if !ok {
		return
	}

	data, err := obj.GetData(r.Context(), r.URL.Query())
	if err != nil {
		h.handleError(w, fmt.Errorf("Pages.GetData: %v '%s'", err, page))
		return
	}

	if data == nil {
		data = make(map[string]any)
	}

	// Навигация (список страниц)
	data["navs"] = h.nav
	data["route"] = page

	if val, ok := data["template"].(string); !ok || val == "" {
		data["template"] = "table"
	}
	if err = h.tmpl.ExecuteTemplate(w, data["template"].(string), data); err != nil {
		h.handleError(w, fmt.Errorf("Pages.ExecuteTemplate: %v '%s', tmpl:%v", err, page, data["template"]))
		return
	}
}

func (h Handler) Form(w http.ResponseWriter, r *http.Request) {
	page, obj, ok := h.checkPageAccess(w, r)
	if !ok {
		return
	}

	var (
		id  int
		err error
	)

	if idStr := r.PathValue("id"); idStr != "" {
		id, err = strconv.Atoi(idStr)
		if err != nil {
			h.attributeError(w, err)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		var data map[string]any
		data, err = obj.GetForm(r.Context(), ds.ID(id), r.URL.Query())
		if err != nil {
			h.handleError(w, err)
			return
		}

		// Навигация (список страниц)
		data["navs"] = h.nav

		if val, ok := data["template"].(string); !ok || val == "" {
			data["template"] = "form"
		}
		if err = h.tmpl.ExecuteTemplate(w, data["template"].(string), data); err != nil {
			h.handleError(w, err)
			return
		}

	case http.MethodPost:
		if err = r.ParseForm(); err != nil {
			h.handleError(w, err)
			return
		}

		if r.Form.Get("_method") == http.MethodDelete {
			if err = obj.Delete(r.Context(), ds.ID(id)); err != nil {
				h.handleError(w, err)
				return
			}
			http.Redirect(w, r, "/"+page, http.StatusFound)
		} else {
			location, err := obj.Upsert(r.Context(), r)
			if err != nil {
				h.handleError(w, err)
				return
			}

			if location != "" {
				if location == "page" {
					http.Redirect(w, r, "/"+page, http.StatusFound)
				} else {
					http.Redirect(w, r, location, http.StatusFound)
				}
			}
		}
	case http.MethodDelete:
		if err = r.ParseForm(); err != nil {
			h.handleError(w, err)
			return
		}

		if err = obj.Delete(r.Context(), ds.ID(id)); err != nil {
			h.handleError(w, err)
			return
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func (h Handler) checkPageAccess(w http.ResponseWriter, r *http.Request) (string, ds.Pager, bool) {
	page := r.PathValue("page")
	var found bool
	for route := range h.registry {
		if string(route) == page {
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return page, nil, false
	}

	pagerFactory := h.FactoryPager(ds.RouteType(page), h.store)
	if pagerFactory == nil {
		w.WriteHeader(http.StatusNotFound)
		return page, nil, false
	}

	return page, pagerFactory, true
}

func (h Handler) attributeError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("ERROR: %q", err)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h Handler) handleError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("ERROR: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
