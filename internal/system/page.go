package system

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

type System struct {
	ID        ds.ID  //`db:"id"`
	ProjectID ds.ID  `db:"project_id" schema:"project_id"`
	Name      string `db:"name" schema:"name"`
	Status    string `db:"status" schema:"status"`
}

func (s System) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("название системы обязательно")
	}
	if s.ProjectID == 0 {
		return fmt.Errorf("project_id обязателен")
	}
	return nil
}

type Page ds.Page

func (obj Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	projectID, _ := strconv.ParseUint(q.Get("project_id"), 10, 64)

	var items []System
	filter := database.SqlFilter{
		Table:   ds.SystemRoute.String(),
		Fields:  "id," + utils.ExtractFields("db", System{}),
		OrderBy: "name",
	}
	if projectID > 0 {
		filter.Where = "project_id = ?"
		filter.Args = []any{projectID}
	}

	if err := obj.Store.Select(ctx, &items, filter); err != nil {
		return nil, err
	}

	list := make([]ds.Row, 0, len(items))
	for _, item := range items {
		list = append(list, ds.Row{
			ID: item.ID,
			Columns: []ds.Column{
				{Type: "text", Val: item.Name},
				{Type: "text", Val: item.Status},
			},
		})
	}

	return map[string]any{
		"route":     obj.Route,
		"pageTitle": "Системы",
		"titles":    []string{"Название", "Статус"},
		"list":      list,
		"projectID": projectID,
		"template":  "systems",
	}, nil
}

func (obj Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	var (
		item System
		err  error
	)

	if id > 0 {
		item, err = GetByID(ctx, obj.Store, id)
		if err != nil {
			return nil, err
		}
	} else {
		item.Status = "новый"
		if pid := q.Get("project_id"); pid != "" {
			if v, err := strconv.ParseUint(pid, 10, 64); err == nil {
				item.ProjectID = ds.ID(v)
			}
		}
	}

	fields := []ds.Input{
		{Label: "Название", Type: "text", Name: "name", Val: item.Name},
		{Label: "Статус", Type: "select", Name: "status", Val: item.Status, Options: statusOptions(item.Status)},
		{Label: "", Type: "integer", Name: "project_id", Val: item.ProjectID, Hidden: true},
	}

	return map[string]any{
		"route":     obj.Route,
		"id":        item.ID,
		"pageTitle": "Система",
		"fields":    fields,
		"projectID": item.ProjectID,
		"template":  "form",
		"colors":    []string{},
	}, nil
}

func (obj Page) Upsert(ctx context.Context, req *http.Request) (string, error) {
	var (
		item System
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

	fields := utils.GetFormFields(req.PostForm, utils.ExtractFields("db", System{}))
	if fields == "" {
		fields = utils.ExtractFields("db", System{})
	}

	_, err = obj.Store.Upsert(ctx, ds.SystemRoute.String(), int64(item.ID), fields, item)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/systems?project_id=%d", item.ProjectID), nil
}

func (obj Page) Delete(ctx context.Context, id ds.ID) error {
	return obj.Store.Delete(ctx, ds.SystemRoute.String(), id)
}

func GetByID(ctx context.Context, store ds.Storage, id ds.ID) (System, error) {
	var items []System
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", System{}),
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	})
	if err != nil {
		return System{}, err
	}
	if len(items) == 0 {
		return System{}, fmt.Errorf("system not found")
	}
	return items[0], nil
}

func GetMap(ctx context.Context, store ds.Storage) (map[ds.ID]System, error) {
	var items []System
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", System{}),
	})
	if err != nil {
		return nil, err
	}
	m := map[ds.ID]System{}
	for i := range items {
		m[items[i].ID] = items[i]
	}
	return m, nil
}

func GetAsSelect(ctx context.Context, store ds.Storage, selected ds.ID) ([]ds.Option, error) {
	var items []System
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:   ds.SystemRoute.String(),
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