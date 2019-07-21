package server

import (
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
)

var bookIndex *template.Template
var bookEdit *template.Template
var bookNew *template.Template
var userRegistrationNew *template.Template

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

	userRegistrationNew, err = loadTemplate("user_registration_new", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "user_registration_new.html")}, RouteFuncMap)
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

func RenderBookIndex(w io.Writer, csrfTemplateTag template.HTML, books []*BooksForYear, username string) error {
	return bookIndex.Execute(w, map[string]interface{}{
		"BooksForYears":  books,
		csrf.TemplateTag: csrfTemplateTag,
		"username":       username,
	})
}

func RenderBookEdit(w io.Writer, csrfTemplateTag template.HTML, bookId int64, uba domain.UpdateBookArgs, verr validate.Errors, username string) error {
	return bookEdit.Execute(w, map[string]interface{}{
		"bookID":         bookId,
		"fields":         uba,
		"errors":         verr,
		csrf.TemplateTag: csrfTemplateTag,
		"username":       username,
	})
}

func RenderBookNew(w io.Writer, csrfTemplateTag template.HTML, cba domain.CreateBookArgs, verr validate.Errors, username string) error {
	return bookNew.Execute(w, map[string]interface{}{
		"fields":         cba,
		"errors":         verr,
		csrf.TemplateTag: csrfTemplateTag,
		"username":       username,
	})
}

func RenderUserRegistrationNew(w io.Writer, csrfTemplateTag template.HTML, urr *UserRegistrationRequest, verr validate.Errors) error {
	return userRegistrationNew.Execute(w, map[string]interface{}{
		"fields":         urr,
		"errors":         verr,
		csrf.TemplateTag: csrfTemplateTag,
	})
}
