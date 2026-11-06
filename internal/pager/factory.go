package pager

import (
	"log"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/internal/product"
	"github.com/SlavaNovikov/grif/internal/project"
	"github.com/SlavaNovikov/grif/internal/report"
	"github.com/SlavaNovikov/grif/internal/system"
	"github.com/SlavaNovikov/grif/internal/user"
	"github.com/SlavaNovikov/grif/internal/work"
)

var Nav = []ds.RouteType{
	ds.ProjectRoute,
	ds.WorkRoute,
	ds.WorksRoute,
	ds.ProductRoute,
	ds.UsersRoute,
	ds.ReportsRoute,
}

var Registry = ds.Registry{
	ds.ProjectRoute: {Nav: ds.ProjectRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return project.Page{Route: r, Store: s} }},
	ds.SystemRoute:  {Nav: ds.SystemRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return system.Page{Route: r, Store: s} }},
	ds.WorkRoute:    {Nav: ds.WorkRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return work.Page{Route: r, Store: s} }},
	ds.WorksRoute:   {Nav: ds.WorksRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return work.WorksPage{Route: r, Store: s} }},
	ds.UsersRoute:   {Nav: ds.UsersRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return user.Page{Route: r, Store: s} }},
	ds.ProductRoute: {Nav: ds.ProductRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return product.Page{Route: r, Store: s} }},
	ds.ReportsRoute: {Nav: ds.ReportsRoute.NavName(), Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return report.Page{Route: r, Store: s} }},
	ds.SearchRoute:  {Nav: "", Factory: func(r ds.RouteType, s ds.Storage) ds.Pager { return product.SearchPage{Route: r, Store: s} }},
}

type Factory func(route ds.RouteType, store ds.Storage) ds.Pager

func (h Handler) FactoryPager(page ds.RouteType, store ds.Storage) ds.Pager {
	info, ok := h.registry[page]
	if !ok {
		log.Printf("ERROR: page not found in registry: %s", page)
		return nil
	}
	return info.Factory(page, store)
}
