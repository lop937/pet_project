package product

import (
	"context"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/SlavaNovikov/grif/internal/ds"
	"net/http"
	"net/url"
	"strings"
)

type SearchPage ds.Page

func (s SearchPage) GetData(ctx context.Context, q url.Values) (map[string]any, error) {
	search := strings.Trim(q.Get("search"), " ")
	page := strings.Trim(q.Get("p"), " ")

	filter := database.SqlFilter{
		Table:  ds.ProductRoute.String(),
		Fields: "id," + utils.ExtractFields("db", Product{}),
		Limit:  100,
	}

	var items []Product

	if page == "work" && search == "" {
		// тут отдадим продукты из последних работ
		var err error
		items, err = GetLastUsedProducts(ctx, s.Store)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			err := s.Store.Select(ctx, &items, filter)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if search != "" {
			filter.Where = "name LIKE ? or tag LIKE ?"
			filter.Args = []any{"%" + search + "%", "%" + search + "%"}
		}
		err := s.Store.Select(ctx, &items, filter)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"route":    page,
		"template": "search_result_" + page,
		"items":    items,
	}, nil
}

func (s SearchPage) GetForm(ctx context.Context, id ds.ID, q url.Values) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (s SearchPage) Upsert(ctx context.Context, r *http.Request) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s SearchPage) Delete(ctx context.Context, id ds.ID) error {
	//TODO implement me
	panic("implement me")
}
