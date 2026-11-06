package project

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

type Project struct {
	ID       ds.ID  //`db:"id"`
	Name     string `db:"name" schema:"name"`
	Date     string `db:"date" schema:"date"`
	Customer string `db:"customer" schema:"customer"`
	Status   string `db:"status" schema:"status"`
}

func (p Project) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("название проекта обязательно")
	}
	return nil
}

type Page ds.Page

func (obj Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	var items []Project

	filter := database.SqlFilter{
		Table:   ds.ProjectRoute.String(),
		Fields:  "id," + utils.ExtractFields("db", Project{}),
		OrderBy: "date DESC",
	}

	from := q.Get("from")
	to := q.Get("to")
	if from != "" || to != "" {
		var where string
		var args []any
		if from != "" {
			where = "date >= ?"
			args = append(args, from)
		}
		if to != "" {
			if where != "" {
				where += " AND "
			}
			where += "date <= ?"
			args = append(args, to)
		}
		filter.Where = where
		filter.Args = args
	}

	if err := obj.Store.Select(ctx, &items, filter); err != nil {
		return nil, err
	}

	list := make([]ds.Row, 0, len(items))
	for _, item := range items {
		list = append(list, ds.Row{
			ID: item.ID,
			Columns: []ds.Column{
				{Type: "date", Val: item.Date},
				{Type: "text", Val: item.Name},
				{Type: "text", Val: item.Customer},
				{Type: "text", Val: item.Status},
			},
		})
	}

	return map[string]any{
		"route":     obj.Route,
		"pageTitle": "Проекты",
		"titles":    []string{"Дата", "Название", "Заказчик", "Статус"},
		"list":      list,
		"template":  "table",
	}, nil
}

func (obj Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	var (
		item Project
		err  error
	)

	if id > 0 {
		item, err = GetByID(ctx, obj.Store, id)
		if err != nil {
			return nil, err
		}
	} else {
		item.Status = "новый"
		item.Date = ds.NowDate()
	}

	fields := []ds.Input{
		{Label: "Название", Type: "text", Name: "name", Val: item.Name},
		{Label: "Дата", Type: "date", Name: "date", Val: item.Date},
		{Label: "Заказчик", Type: "text", Name: "customer", Val: item.Customer},
		{Label: "Статус", Type: "select", Name: "status", Val: item.Status, Options: statusOptions(item.Status)},
	}

	return map[string]any{
		"route":     obj.Route,
		"id":        item.ID,
		"pageTitle": "Проект",
		"fields":    fields,
		"template":  "project_form",
		"colors":    []string{},
		"showSystems": item.ID > 0,
		"systemsURL":  fmt.Sprintf("/systems?project_id=%d", item.ID),
	}, nil
}

func (obj Page) Upsert(ctx context.Context, req *http.Request) (string, error) {
	var (
		item Project
		err  error
	)

	schemaDecoder.IgnoreUnknownKeys(true)
	if err = schemaDecoder.Decode(&item, req.PostForm); err != nil {
		return "", err
	}

	item.ID, err = ds.IDFromReq(req)
	if err != nil {
		return "", err
	}

	if err = item.Validate(); err != nil {
		return "", err
	}

	if item.Date == "" {
		item.Date = ds.NowDate()
	}

	fields := utils.GetFormFields(req.PostForm, utils.ExtractFields("db", Project{}))
	if fields == "" {
		fields = utils.ExtractFields("db", Project{})
	}

	_, err = obj.Store.Upsert(ctx, ds.ProjectRoute.String(), int64(item.ID), fields, item)
	return "page", err
}

func (obj Page) Delete(ctx context.Context, id ds.ID) error {
	return obj.Store.Delete(ctx, ds.ProjectRoute.String(), id)
}

func GetByID(ctx context.Context, store ds.Storage, id ds.ID) (Project, error) {
	var items []Project
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.ProjectRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Project{}),
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	})
	if err != nil {
		return Project{}, err
	}
	if len(items) == 0 {
		return Project{}, fmt.Errorf("project not found")
	}
	return items[0], nil
}

func GetMap(ctx context.Context, store ds.Storage) (map[ds.ID]Project, error) {
	var items []Project
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.ProjectRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Project{}),
	})
	if err != nil {
		return nil, err
	}
	m := map[ds.ID]Project{}
	for i := range items {
		m[items[i].ID] = items[i]
	}
	return m, nil
}

func GetAsSelect(ctx context.Context, store ds.Storage, selected ds.ID) ([]ds.Option, error) {
	var items []Project
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:   ds.ProjectRoute.String(),
		Fields:  "id,name",
		OrderBy: "name",
	})
	if err != nil {
		return nil, err
	}
	options := []ds.Option{}
	for i := range items {
		options = append(options, ds.Option{
			Selected: items[i].ID == selected,
			Val:      items[i].ID,
			Name:     items[i].Name,
		})
	}
	return options, nil
}

func statusOptions(selected string) []ds.Option {
	values := []string{"новый", "в работе", "завершён"}
	options := make([]ds.Option, 0, len(values))
	for _, v := range values {
		options = append(options, ds.Option{
			Selected: v == selected,
			Val:      v,
			Name:     v,
		})
	}
	return options
}