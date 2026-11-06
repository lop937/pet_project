package product

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/gorilla/schema"
)

type Product struct {
	ID     ds.ID
	Tag    string  `db:"tag" schema:"tag"`
	Name   string  `db:"name" schema:"name"`
	Price  float64 `db:"price" schema:"price"`
	Unit   string  `db:"unit" schema:"unit"`
	Square float64 `db:"square" schema:"square"`
	Attr   string  `db:"attr" schema:"attr"`
}

type (
	Page ds.Page
	Type int
)

var schemaDecoder = schema.NewDecoder()

func (obj Product) Validate() error {
	return nil
}

func (obj Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	return map[string]any{
		"route":    obj.Route,
		"template": "products",
	}, nil
}

func (obj Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	var item Product
	var err error

	if id > 0 {
		item, err = GetByID(ctx, obj.Store, id)
		if err != nil {
			return nil, err
		}
	}

	if item.Unit == "" {
		item.Unit = "метры в 2"
	}

	return map[string]any{
		"route":     obj.Route,
		"id":        item.ID,
		"pageTitle": "",

		"template": "product_form",

		"fields": []ds.Input{
			{Label: "Название", Type: "text", Name: "name", Val: item.Name},
			{Label: "Единица измерения", Type: "select", Name: "unit", Val: item.Unit, Options: unitOptions(item.Unit)},
			// Подписи для этих полей будут меняться через JS
			{Label: "Рубли за метр квадратный", Type: "text", Name: "price", Val: item.Price, Attrs: `inputmode="decimal"`},
			{Label: "Метров квадратных в единице", Type: "text", Name: "square", Val: item.Square, Attrs: `inputmode="decimal"`},
		},
		"colors": []string{},
	}, nil
}

func (obj Page) Upsert(ctx context.Context, req *http.Request) (string, error) {
	var (
		item Product
		err  error
	)

	if err := req.ParseForm(); err != nil {
		return "", err
	}

	// Заменяем запятые на точки для всех числовых полей формы
	for _, key := range []string{"price", "square"} {
		if v := req.PostForm.Get(key); v != "" {
			req.PostForm.Set(key, strings.ReplaceAll(v, ",", "."))
		} else {
			// Если поле пустое — явно ставим "0", чтобы gorilla/schema не падала
			req.PostForm.Set(key, "0")
		}
	}

	// Декодируем форму в структуру
	if err = schemaDecoder.Decode(&item, req.PostForm); err != nil {
		return "", fmt.Errorf("ошибка декодирования формы: %w", err)
	}

	item.ID, err = ds.IDFromReq(req)
	if err != nil {
		return "", err
	}

	// Сохраняем только те поля, которые пришли из формы (tag и attr не трогаем)
	fields := utils.GetFormFields(req.PostForm, utils.ExtractFields("db", Product{}))
	if fields == "" {
		fields = utils.ExtractFields("db", Product{})
	}

	if err = item.Validate(); err != nil {
		return "", err
	}

	_, err = obj.Store.Upsert(ctx, obj.Route.String(), int64(item.ID), fields, item)
	return "page", err
}

func (obj Page) Delete(ctx context.Context, id ds.ID) error {
	return obj.Store.Delete(ctx, obj.Route.String(), id)
}

func GetByID(ctx context.Context, store ds.Storage, id ds.ID) (Product, error) {
	var items []Product
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.ProductRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Product{}),
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	})
	if err != nil {
		return Product{}, err
	}
	if len(items) == 0 {
		return Product{}, fmt.Errorf("product not found")
	}
	return items[0], nil
}

func GetMap(ctx context.Context, store ds.Storage) (map[ds.ID]Product, error) {
	var items []Product
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:  ds.ProductRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Product{}),
	})
	if err != nil {
		return nil, err
	}

	productMap := map[ds.ID]Product{}
	for i := range items {
		productMap[items[i].ID] = items[i]
	}
	return productMap, nil
}

func GetLastUsedProducts(ctx context.Context, store ds.Storage) ([]Product, error) {
	var products []Product
	err := store.SelectX(ctx, &products, `
		SELECT p.*
		FROM products p
		JOIN (
			SELECT product_id, MAX(id) as last_work_id
			FROM works
			GROUP BY product_id
			ORDER BY last_work_id DESC
			LIMIT 20
		) last_items ON p.id = last_items.product_id 
		ORDER BY last_items.last_work_id DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении последних продуктов: %w", err)
	}
	return products, nil
}

// unitOptions возвращает список единиц измерения
func unitOptions(selected string) []ds.Option {
	values := []string{"метры в 2", "метры погонные", "штуки"}
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

// GetAsSelect возвращает список продуктов для выпадающего списка (используется в других пакетах)
func GetAsSelect(ctx context.Context, store ds.Storage, selected ds.ID) ([]ds.Option, error) {
	var items []Product
	err := store.Select(ctx, &items, database.SqlFilter{
		Table:   ds.ProductRoute.String(),
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
