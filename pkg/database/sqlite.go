package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DB — обёртка над sqlx.DB.
type DB struct {
	*sqlx.DB
}

// NewSQLite открывает SQLite-базу и выполняет initSQL (создание таблиц).
func NewSQLite(path string, initSQL string) *DB {
	db, err := sqlx.Open("sqlite", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		panic(fmt.Errorf("database.NewSQLite: open: %w", err))
	}
	if err = db.Ping(); err != nil {
		panic(fmt.Errorf("database.NewSQLite: ping: %w", err))
	}
	if initSQL != "" {
		if _, err = db.Exec(initSQL); err != nil {
			panic(fmt.Errorf("database.NewSQLite: init: %w", err))
		}
	}
	return &DB{db}
}

// Select собирает SQL из SqlFilter и сканирует результат в dest (указатель на слайс).
func (db *DB) Select(ctx context.Context, dest interface{}, f SqlFilter) error {
	if f.Table == "" {
		return fmt.Errorf("database.Select: empty table")
	}
	fields := f.Fields
	if fields == "" {
		fields = "*"
	}
	query := "SELECT " + fields + " FROM " + f.Table

	args := make([]any, 0, len(f.Args))

	if f.Where != nil {
		switch w := f.Where.(type) {
		case string:
			if strings.TrimSpace(w) != "" {
				query += " WHERE " + w
				args = append(args, f.Args...)
			}
		default:
			return fmt.Errorf("database.Select: unsupported Where type %T", f.Where)
		}
	}

	if f.OrderBy != "" {
		query += " ORDER BY " + f.OrderBy
	}

	if f.Limit != nil {
		switch l := f.Limit.(type) {
		case string:
			if strings.TrimSpace(l) != "" {
				query += " LIMIT " + l
			}
		case int:
			query += fmt.Sprintf(" LIMIT %d", l)
		case int64:
			query += fmt.Sprintf(" LIMIT %d", l)
		default:
			return fmt.Errorf("database.Select: unsupported Limit type %T", f.Limit)
		}
	}

	return db.DB.SelectContext(ctx, dest, query, args...)
}

// SelectX выполняет сырой SQL и сканирует результат в dest.
func (db *DB) SelectX(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.DB.SelectContext(ctx, dest, query, args...)
}

// Upsert: id == 0 → INSERT, иначе UPDATE.
// fields — список колонок через запятую (например "tag,name,price").
// data — структура, из которой значения берутся по тегу `db`.
func (db *DB) Upsert(ctx context.Context, table string, id int64, fields string, data interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("database.Upsert: empty table")
	}
	cols := splitFields(fields)
	if len(cols) == 0 {
		return 0, fmt.Errorf("database.Upsert: empty fields")
	}

	values, err := structValuesByTags(data, cols, "db")
	if err != nil {
		return 0, err
	}

	if id == 0 {
		placeholders := make([]string, len(cols))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			table, strings.Join(cols, ","), strings.Join(placeholders, ","),
		)
		res, err := db.DB.ExecContext(ctx, query, values...)
		if err != nil {
			return 0, fmt.Errorf("database.Upsert insert: %w", err)
		}
		return res.LastInsertId()
	}

	setParts := make([]string, len(cols))
	for i, c := range cols {
		setParts[i] = c + "=?"
	}
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id=?",
		table, strings.Join(setParts, ","),
	)
	args := append(values, id)
	if _, err = db.DB.ExecContext(ctx, query, args...); err != nil {
		return 0, fmt.Errorf("database.Upsert update: %w", err)
	}
	return id, nil
}

// Delete удаляет строку по id.
func (db *DB) Delete(ctx context.Context, table string, id any) error {
	if table == "" {
		return fmt.Errorf("database.Delete: empty table")
	}
	_, err := db.DB.ExecContext(ctx, "DELETE FROM "+table+" WHERE id=?", id)
	return err
}

// ---------- helpers ----------

func splitFields(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// structValuesByTags возвращает значения полей структуры в порядке cols,
// беря имена из тега tagName. Поддерживает указатель на структуру.
func structValuesByTags(data interface{}, cols []string, tagName string) ([]any, error) {
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("database: nil pointer")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("database: expected struct, got %s", v.Kind())
	}
	t := v.Type()

	// map: tag -> field index
	tagToIdx := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}
		// поддерживаем "db:\"name,omitempty\"" — берём первое
		if idx := strings.Index(tag, ","); idx >= 0 {
			tag = tag[:idx]
		}
		tagToIdx[tag] = i
	}

	out := make([]any, 0, len(cols))
	for _, c := range cols {
		idx, ok := tagToIdx[c]
		if !ok {
			return nil, fmt.Errorf("database: field with tag %q=%q not found", tagName, c)
		}
		out = append(out, v.Field(idx).Interface())
	}
	return out, nil
}

// гарантия, что sql.ErrNoRows доступен, если понадобится
var _ = sql.ErrNoRows
