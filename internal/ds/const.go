package ds

import "time"

// Все route в системе
const (
	ProjectRoute RouteType = "projects"
	SystemRoute  RouteType = "systems"
	WorkRoute    RouteType = "work"
	WorksRoute   RouteType = "works"
	UsersRoute   RouteType = "users"
	ProductRoute RouteType = "products"
	ReportsRoute RouteType = "reports"
	SearchRoute  RouteType = "search"
)

const (
	LayoutEventDatetime = "2006-01-02T15:04"
	LayoutEventDate     = "2006-01-02"
	LayoutEventTime     = "15:04"
)

// NowDate возвращает текущую дату в формате LayoutEventDate.
func NowDate() string {
	return time.Now().Format(LayoutEventDate)
}

func (r RouteType) NavName() string {
	switch r {
	case ProjectRoute:
		return "Проекты"
	case SystemRoute:
		return "Системы"
	case WorkRoute:
		return "Работа"
	case WorksRoute:
		return "Учет"
	case ProductRoute:
		return "Каталог"
	case UsersRoute:
		return "Люди"
	case ReportsRoute:
		return "Сводка"
	default:
		return string(r)
	}
}