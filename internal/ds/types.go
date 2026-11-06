package ds

import (
	"context"
	"github.com/SlavaNovikov/grif/pkg/database"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Storage interface {
	Select(ctx context.Context, dest interface{}, filter database.SqlFilter) error
	Upsert(ctx context.Context, table string, id int64, fields string, data interface{}) (int64, error)
	Delete(ctx context.Context, table string, id any) error
	SelectX(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Close() error
}

// Pager — интерфейс, который должны реализовывать все страницы
type Pager interface {
	GetData(ctx context.Context, q url.Values) (map[string]any, error)
	GetForm(ctx context.Context, id ID, q url.Values) (map[string]any, error)
	Upsert(ctx context.Context, r *http.Request) (string, error)
	Delete(ctx context.Context, id ID) error
}

// Factory — сигнатура функции-конструктора для создания страницы
type Factory func(route RouteType, store Storage) Pager

// PageInfo объединяет навигацию в меню и конструктор для каждой страницы
type PageInfo struct {
	Nav     string
	Factory Factory
}

type (
	ID        uint64
	PowerStr  string
	RouteType string
	Registry  map[RouteType]PageInfo
)

type (
	Group struct {
		ID    ID
		Title string
		Items []EventSlot
	}
	EventSlot struct {
		ID        ID
		Timestamp int64
		Start     int64
		StartDate string
		StartTime string
		EndTime   string
		Name      string
		Trainer   string
		Color     string
		Duration  int
		Height    int
		Offset    int
		Meta      string
		Rec       int
		Visits    int
	}
)

type (
	Row struct {
		ID      ID
		Columns []Column
	}
	Column struct {
		Val     any
		Type    string
		Events  []EventSlot
		Disable bool
	}
)

type (
	Form struct {
		ID     ID
		Fields []Input
	}
	Input struct {
		Label    string
		Type     string
		Name     string
		Val      any
		Attrs    string
		Selected string
		Options  []Option
		Group    []Input
		Disabled bool
		Hidden   bool
	}
	Option struct {
		Val      any
		Name     string
		Color    string
		Selected bool
	}
)

func (p PowerStr) Array() []string {
	if len(p) == 0 {
		return nil
	}
	return strings.Split(string(p), ",")
}

func (p PowerStr) ToString(v []string) PowerStr {
	return PowerStr(strings.Join(v, ","))
}

func (r RouteType) String() string {
	return string(r)
}

func IDFromReq(r *http.Request) (ID, error) {
	if idStr := r.PathValue("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, err
		}
		return ID(id), nil
	}

	return 0, nil
}

type Page struct {
	Route RouteType
	Store Storage
}


