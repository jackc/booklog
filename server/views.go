package server

import (
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/gorilla/csrf"
)

var bookIndex *template.Template
var bookEdit *template.Template
var bookNew *template.Template

func LoadTemplates(path string) error {
	var err error

	bookIndex, err = loadTemplate("book_index", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_index.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookEdit, err = loadTemplate("book_edit", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_edit.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookNew, err = loadTemplate("book_new", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_new.html")}, RouteFuncMap)
	if err != nil {
		return err
	}
	return nil
}

func loadTemplate(name string, files []string, funcMap template.FuncMap) (*template.Template, error) {
	tmpl := template.New(name)
	tmpl.Funcs(funcMap)

	for _, file := range files {
		src, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(string(src))
		if err != nil {
			return nil, err
		}

	}

	return tmpl, nil
}

func RenderBookIndex(w io.Writer, csrfTemplateTag template.HTML, books []BookRow001) error {
	return bookIndex.Execute(w, map[string]interface{}{
		"Books":          books,
		csrf.TemplateTag: csrfTemplateTag,
	})
}

func RenderBookEdit(w io.Writer, csrfTemplateTag template.HTML, bookId string, bcr *BookCreateRequest, errors interface{}) error {
	return bookEdit.Execute(w, map[string]interface{}{
		"bookID":         bookId,
		"fields":         bcr,
		"errors":         errors,
		csrf.TemplateTag: csrfTemplateTag,
	})
}

func RenderBookNew(w io.Writer, csrfTemplateTag template.HTML, bcr *BookCreateRequest, errors interface{}) error {
	return bookNew.Execute(w, map[string]interface{}{
		"fields":         bcr,
		"errors":         errors,
		csrf.TemplateTag: csrfTemplateTag,
	})
}
