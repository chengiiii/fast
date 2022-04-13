package server

import (
	"embed"
	"fast/fileservice"
	"html/template"
)

var rootTemplate *template.Template

//go:embed templates/*
var f embed.FS

func ImportTemplates() error {
	var err error
	rootTemplate, err = template.ParseFS(f,
		"templates/*.gohtml",
		"templates/static/*.js",
		"templates/static/*.css")
	if err != nil {
		return err
	}
	return nil
}

type ViewData struct {
	Files     fileservice.FileList
	Upload    bool
	Style     string
	Script    string
	JsonFiles string
}

func getCSS() string {
	b, err := f.ReadFile("templates/static/index.css")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func getScript() string {
	b, err := f.ReadFile("templates/static/index.js")
	if err != nil {
		panic(err)
	}
	return string(b)
}
