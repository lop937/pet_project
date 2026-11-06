package work

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/internal/product"
	"github.com/SlavaNovikov/grif/internal/project"
	"github.com/SlavaNovikov/grif/internal/report"
	"github.com/SlavaNovikov/grif/internal/system"
	"github.com/SlavaNovikov/grif/internal/user"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

type Work struct {
	ID          ds.ID   //`db:"id"`
	Date        string  `db:"date" schema:"date"`
	Name        string  `db:"name" schema:"name"`
	Price       float64 `db:"price" schema:"price"`
	UserID      ds.ID   `db:"user_id" schema:"user_id"`
	ProductID   ds.ID   `db:"product_id" schema:"product_id"`
	ProjectID   ds.ID   `db:"project_id" schema:"project_id"`
	SystemID    ds.ID   `db:"system_id" schema:"system_id"`
	ServiceType string  `db:"service_type" schema:"service_type"`
	Quantity    float64 `db:"quantity" schema:"quantity"`
}

func (w Work) Validate() error {
	if w.ProductID == 0 {
		return fmt.Errorf("товар не выбран")
	}
	if w.ProjectID == 0 {
		return fmt.Errorf("выберите проект")
	}
	if w.SystemID == 0 {
		return fmt.Errorf("выберите систему")
	}
	if w.UserID == 0 {
		return fmt.Errorf("выберите исполнителя")
	}
	return nil
}

type Page ds.Page

// GetData — страница /work (форма быстрого ввода через поиск)
func (p Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	dayReport, err := report.GetSummary(ctx, p.Store, "", "")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"template":   "work",
		"day_report": dayReport,
	}, nil
}

// GetForm — форма создания работы (открывается из /work после выбора продукта)
// GetForm — форма создания работы (открывается из /work после выбора продукта)
func (p Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	users, err := user.GetUsersAsSelect(ctx, p.Store, 0)
	if err != nil {
		return nil, err
	}

	projects, err := project.GetAsSelect(ctx, p.Store, 0)
	if err != nil {
		return nil, err
	}

	systems, err := system.GetAsSelect(ctx, p.Store, 0)
	if err != nil {
		return nil, err
	}

	productID, err := strconv.Atoi(q.Get("product"))
	if err != nil {
		return nil, err
	}

	pr, err := product.GetByID(ctx, p.Store, ds.ID(productID))
	if err != nil {
		return nil, err
	}

	fields := []ds.Input{
		{Label: "", Type: "integer", Name: "product_id", Val: productID, Hidden: true},
		{Label: "Название", Type: "text", Name: "name", Val: pr.Name, Disabled: true},
		{Label: "Проект", Type: "select", Name: "project_id", Val: 0, Options: projects},
		{Label: "Система", Type: "select", Name: "system_id", Val: 0, Options: systems},
		{Label: "Услуга", Type: "select", Name: "service_type", Val: "монтаж", Options: serviceOptions("монтаж")},
		{Label: "Исполнитель", Type: "select", Name: "user_id", Val: 0, Options: users},
		{Label: "", Type: "group", Name: "", Group: []ds.Input{
			{Label: "Кол-во", Type: "number", Name: "quantity", Val: 1},
			// ВОТ ТУТ МЕНЯЕМ НА pr.Price:
			{Label: "Цена", Type: "number", Name: "price", Val: pr.Price},
		}},
	}

	return map[string]any{
		"id":      0,
		"route":   ds.WorksRoute,
		"product": pr,
		"fields":  fields,
		"colors":  []string{},
	}, nil
}

func (p Page) Upsert(ctx context.Context, r *http.Request) (string, error) {
	return "", nil
}

func (p Page) Delete(ctx context.Context, id ds.ID) error {
	return nil
}

// ============ WorksPage — список работ и редактирование ============

type WorksPage ds.Page

func (p WorksPage) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	var items []Work

	if err := p.Store.Select(ctx, &items, database.SqlFilter{
		Table:   ds.WorksRoute.String(),
		Fields:  "id," + utils.ExtractFields("db", Work{}),
		OrderBy: "date DESC, id DESC",
	}); err != nil {
		return nil, err
	}

	usersMap, err := user.GetUsersMap(ctx, p.Store)
	if err != nil {
		return nil, err
	}

	productMap, err := product.GetMap(ctx, p.Store)
	if err != nil {
		return nil, err
	}

	projectMap, err := project.GetMap(ctx, p.Store)
	if err != nil {
		return nil, err
	}

	systemMap, err := system.GetMap(ctx, p.Store)
	if err != nil {
		return nil, err
	}

	list := make([]ds.Row, 0, len(items))

	for _, item := range items {
		projectName := ""
		if pr, ok := projectMap[item.ProjectID]; ok {
			projectName = pr.Name
		}

		systemName := ""
		if sys, ok := systemMap[item.SystemID]; ok {
			systemName = sys.Name
		}

		list = append(list, ds.Row{
			ID: item.ID,
			Columns: []ds.Column{
				{Type: "checkbox", Val: item.ID},
				{Type: "date", Val: item.Date},
				{Type: "text", Val: projectName},
				{Type: "text", Val: systemName},
				{Type: "text", Val: productMap[item.ProductID].Name},
				{Type: "text", Val: item.ServiceType},
				{Type: "text", Val: usersMap[item.UserID].FIO},
				{Type: "number", Val: item.Quantity},
				{Type: "number", Val: item.Quantity * item.Price}, // Показываем итоговую сумму за кол-во,
			},
		})
	}

	dayReport, err := report.GetSummary(ctx, p.Store, "", "")
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"route":      p.Route,
		"pageTitle":  "",
		"titles":     []string{"Дата", "Проект", "Система", "Товар", "Услуга", "Исполнитель", "Кол-во", "Сумма"},
		"list":       list,
		"day_report": dayReport,
	}, nil
}

func (p WorksPage) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	var (
		item Work
		err  error
	)

	if id > 0 {
		if item, err = selectWorkByID(ctx, p.Store, id); err != nil {
			return nil, err
		}
	}

	// Списки для select'ов
	users, err := user.GetUsersAsSelect(ctx, p.Store, item.UserID)
	if err != nil {
		return nil, err
	}
	users = append([]ds.Option{{Name: "", Val: 0}}, users...)

	projects, err := project.GetAsSelect(ctx, p.Store, item.ProjectID)
	if err != nil {
		return nil, err
	}

	systems, err := system.GetAsSelect(ctx, p.Store, item.SystemID)
	if err != nil {
		return nil, err
	}

	// Продукт (название)
	productName := ""
	var pr product.Product // Объявляем переменную ЗАРАНЕЕ

	if item.ProductID > 0 {
		if p, err := product.GetByID(ctx, p.Store, item.ProductID); err == nil {
			productName = p.Name
			pr = p // Присваиваем значение
		}
	}

	if item.ServiceType == "" {
		item.ServiceType = "монтаж"
	}

	fields := []ds.Input{
		{Label: "Дата", Type: "date", Name: "date", Val: item.Date},
		{Label: "Проект", Type: "select", Name: "project_id", Val: item.ProjectID, Options: projects},
		{Label: "Система", Type: "select", Name: "system_id", Val: item.SystemID, Options: systems},
		{Label: "Товар", Type: "text", Name: "", Val: productName, Disabled: true},
		{Label: "Услуга", Type: "select", Name: "service_type", Val: item.ServiceType, Options: serviceOptions(item.ServiceType)},
		{Label: "Исполнитель", Type: "select", Name: "user_id", Val: item.UserID, Options: users},
		{Label: "", Type: "group", Name: "", Group: []ds.Input{
			{Label: "Кол-во", Type: "number", Name: "quantity", Val: item.Quantity},
			{Label: "Цена", Type: "number", Name: "price", Val: item.Price},
		}},
		{Label: "", Type: "integer", Name: "product_id", Val: item.ProductID, Hidden: true},
	}

	return map[string]any{
		"id":         item.ID,
		"route":      p.Route,
		"product":    pr,             // <--- ДОБАВЛЯЕМ ВОТ ЭТУ СТРОКУ
		"project_id": item.ProjectID, // <--- ДОБАВИТЬ ВОТ ЭТУ СТРОКУ
		"fields":     fields,
		"colors":     []string{},
	}, nil
}

func (p WorksPage) Upsert(ctx context.Context, req *http.Request) (string, error) {
	var (
		item Work
		err  error
	)

	if err := req.ParseForm(); err != nil {
		return "", err
	}

	// 1. Сначала декодируем форму в структуру (тут цена станет 450, как в форме)
	schemaDecoder.IgnoreUnknownKeys(true)
	if err = schemaDecoder.Decode(&item, req.PostForm); err != nil {
		return "", err
	}

	// 2. Затем подтягиваем название и ПЕРЕСЧИТЫВАЕМ ЦЕНУ
	if item.ProductID > 0 {
		if pr, err := product.GetByID(ctx, p.Store, item.ProductID); err == nil {
			item.Name = pr.Name // Всегда сохраняем актуальное название

			// Если у товара есть площадь (например 2.55), умножаем цену на неё
			if pr.Square > 0 {
				item.Price = pr.Price * pr.Square // В базу уйдет 1147.5
			}
		}
	}

	item.ID, err = ds.IDFromReq(req)
	if err != nil {
		return "", err
	}

	if item.Date == "" {
		item.Date = ds.NowDate()
	}
	if item.ServiceType == "" {
		item.ServiceType = "монтаж"
	}

	if err = item.Validate(); err != nil {
		return "", err
	}

	fields := utils.ExtractFields("db", Work{})

	_, err = p.Store.Upsert(ctx, ds.WorksRoute.String(), int64(item.ID), fields, item)

	return "/works", err
}

func (p WorksPage) Delete(ctx context.Context, id ds.ID) error {
	return p.Store.Delete(ctx, ds.WorksRoute.String(), id)
}

func selectWorkByID(ctx context.Context, store ds.Storage, id ds.ID) (Work, error) {
	var dst []Work
	if err := store.Select(ctx, &dst, database.SqlFilter{
		Table:  ds.WorksRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Work{}),
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	}); err != nil {
		return Work{}, err
	}
	if len(dst) == 0 {
		return Work{}, fmt.Errorf("work not found")
	}
	return dst[0], nil
}

func serviceOptions(selected string) []ds.Option {
	values := []string{"монтаж", "демонтаж"}
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
