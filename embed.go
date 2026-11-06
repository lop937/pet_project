package grif

import "embed"

//go:embed templates/*.gohtml
var TemplatesFS embed.FS

//go:embed static/*
var StaticFS embed.FS
