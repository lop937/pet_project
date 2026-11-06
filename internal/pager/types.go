package pager

import (
	"github.com/SlavaNovikov/grif/internal/ds"
)

type Page struct {
	title string
	route string
	store ds.Storage
}
