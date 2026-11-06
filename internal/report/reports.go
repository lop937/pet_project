package report

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SlavaNovikov/grif/internal/ds"
)

type Page ds.Page

type Summary struct {
	Amount        float64 `db:"amount"`         // Сумма в копейках
	TotalWorks    int64   `db:"total_works"`    // Количество записей
	TotalQuantity float64 `db:"total_quantity"` // Объем (метры, кг и т.д.)
}

func (r Page) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	summary, err := GetSummary(ctx, r.Store, q.Get("f"), q.Get("t"))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"template": "report",
		"data":     summary,
	}, nil
}

func (r Page) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (r Page) Upsert(ctx context.Context, req *http.Request) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (r Page) Delete(ctx context.Context, id ds.ID) error {
	//TODO implement me
	panic("implement me")
}

func GetSummary(ctx context.Context, store ds.Storage, from, to string) (Summary, error) {
	var results []Summary

	if from == "" {
		from = time.Now().Format(ds.LayoutEventDate)
	}
	if to == "" {
		to = from // Если конца периода нет, считаем за один день (from)
	}

	var data Summary
	// 3. Выполнение запроса
	if err := store.SelectX(ctx, &results, `
        SELECT IFNULL(SUM(quantity * price), 0) AS amount,
               COUNT(id)                        AS total_works,
               IFNULL(SUM(quantity), 0)         AS total_quantity
        FROM works
        WHERE date BETWEEN ? AND ?;
    `, from, to); err != nil {
		return data, fmt.Errorf("summary report error: %w", err)
	}

	if len(results) > 0 {
		data = results[0]
	}

	return data, nil
}
