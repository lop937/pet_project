package user

import (
	"context"
	"fmt"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/gorilla/schema"
	"net/http"
	"net/url"

	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/SlavaNovikov/grif/internal/ds"
)

type User struct {
	ID     ds.ID
	FIO    string  `db:"fio" schema:"fio"`
	Rate   float64 `db:"rate" schema:"rate"`     // Коэффициент ставки
	Active bool    `db:"active" schema:"active"` // Добавляем флаг активности вместо удаления
}

var schemaDecoder = schema.NewDecoder()

func (u User) Validate() error {
	return nil
}

type Page ds.Page

func (obj Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	var items []User
	if err := SelectUsers(ctx, obj.Store, &items, database.SqlFilter{}); err != nil {
		return nil, err
	}

	list := make([]ds.Row, 0, len(items))

	for _, item := range items {
		list = append(list, ds.Row{
			ID: item.ID,
			Columns: []ds.Column{
				{Val: item.FIO},
				{Type: "number", Val: item.Rate},
			},
		})
	}

	return map[string]any{
		"route":     obj.Route,
		"pageTitle": "",
		"titles":    []string{"ФИО", "Коэфф."},
		"list":      list,
	}, nil
}

func (obj Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	var (
		item User
		err  error
	)

	var defaultInputs []ds.Input
	if id > 0 {
		if item, err = selectUserByID(ctx, obj.Store, id); err != nil {
			return nil, err
		}
	}

	defaultInputs = append(defaultInputs, []ds.Input{
		{Label: "ФИО", Type: "text", Name: "fio", Val: item.FIO},
		{Label: "Коэффициент", Type: "integer", Name: "rate", Val: item.Rate},
	}...)

	return map[string]any{
		"route":     obj.Route,
		"id":        item.ID,
		"pageTitle": "",
		"fields":    defaultInputs,
	}, nil
}

func (obj Page) Upsert(ctx context.Context, req *http.Request) (string, error) {
	var (
		item User
		err  error
	)
	schemaDecoder.IgnoreUnknownKeys(true)
	if err = schemaDecoder.Decode(&item, req.PostForm); err != nil {
		return "", err
	}

	fields := utils.GetFormFields(req.PostForm, utils.ExtractFields("db", User{}))

	item.ID, err = ds.IDFromReq(req)
	if err != nil {
		return "", err
	}

	if err = item.Validate(); err != nil {
		return "", err
	}
	_, err = obj.Store.Upsert(ctx, ds.UsersRoute.String(), int64(item.ID), fields, item)

	return "page", err
}

func (obj Page) Delete(ctx context.Context, id ds.ID) error {
	// нельзя удалять самого себя
	// и только тех к кому есть доступ
	return obj.Store.Delete(ctx, ds.UsersRoute.String(), id)
}

func selectUserByID(ctx context.Context, store ds.Storage, id ds.ID) (User, error) {
	var dst []User
	if err := store.Select(ctx, &dst, database.SqlFilter{
		Table:  ds.UsersRoute.String(),
		Fields: "id," + utils.ExtractFields("db", User{}),
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	}); err != nil {
		return User{}, err
	}

	if len(dst) == 0 {
		return User{}, fmt.Errorf("user not found")
	}

	return dst[0], nil
}

func SelectUsers(ctx context.Context, store ds.Storage, dst interface{}, filter database.SqlFilter) error {
	filter.Table = ds.UsersRoute.String()
	filter.Fields = "id," + utils.ExtractFields("db", User{})

	if err := store.Select(ctx, dst, filter); err != nil {
		return err
	}

	return nil
}

func GetUsersAsSelect(ctx context.Context, store ds.Storage, id ds.ID) ([]ds.Option, error) {
	var items []User
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.UsersRoute.String(),
		Fields: "id,fio",
	})
	if err != nil {
		return nil, fmt.Errorf("%q, %s", err, ds.UsersRoute)
	}
	var options []ds.Option
	for i := range items {
		options = append(options, ds.Option{
			Selected: items[i].ID == id,
			Val:      items[i].ID,
			Name:     items[i].FIO,
		})
	}

	return options, nil
}

func GetUsersMap(ctx context.Context, store ds.Storage) (map[ds.ID]User, error) {
	var trainers []User
	if err := SelectUsers(ctx, store, &trainers, database.SqlFilter{}); err != nil {
		return nil, err
	}

	trainersMap := map[ds.ID]User{}
	for i := range trainers {
		trainersMap[trainers[i].ID] = trainers[i]
	}

	return trainersMap, nil
}
