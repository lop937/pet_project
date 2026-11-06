package database

// SqlFilter описывает простой фильтр для Select.
type SqlFilter struct {
	Table   string
	Fields  string
	Where   any
	Limit   any
	OrderBy string
	Args    []any
}
