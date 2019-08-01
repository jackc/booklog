package server

import (
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
)

var bookEdit *template.Template
var bookShow *template.Template
var bookConfirmDelete *template.Template
var bookNew *template.Template
var bookImportCSVForm *template.Template

func LoadTemplates(path string) error {
	var err error

	bookEdit, err = loadTemplate("book_edit", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_edit.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookShow, err = loadTemplate("book_show", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_show.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookConfirmDelete, err = loadTemplate("book_confirm_delete", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_confirm_delete.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookNew, err = loadTemplate("book_new", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_new.html")}, RouteFuncMap)
	if err != nil {
		return err
	}

	bookImportCSVForm, err = loadTemplate("book_import_csv_form", []string{filepath.Join(path, "layout.html"), filepath.Join(path, "book_import_csv_form.html")}, RouteFuncMap)
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

type baseViewData struct {
	csrfTemplateTag template.HTML
	session         *Session
}

func baseViewDataFromRequest(r *http.Request) baseViewData {
	return baseViewData{
		csrfTemplateTag: csrf.TemplateField(r),
		session:         r.Context().Value(RequestSessionKey).(*Session),
	}
}

func baseViewArgsFromRequest(r *http.Request) *view.BaseViewArgs {
	var currentUser *data.UserMin
	if um, ok := r.Context().Value(RequestPathUserKey).(*data.UserMin); ok {
		currentUser = um
	}

	return &view.BaseViewArgs{
		CSRFField:   string(csrf.TemplateField(r)),
		CurrentUser: &r.Context().Value(RequestSessionKey).(*Session).User,
		PathUser:    currentUser,
	}
}

func RenderBookEdit(w io.Writer, b baseViewData, bookId int64, form BookEditForm, verr validate.Errors, username string) error {
	return bookEdit.Execute(w, map[string]interface{}{
		"bookID":         bookId,
		"fields":         form,
		"errors":         verr,
		csrf.TemplateTag: b.csrfTemplateTag,
		"session":        b.session,
		"username":       username,
	})
}

func RenderBookShow(w io.Writer, b baseViewData, book *data.Book, username string) error {
	return bookShow.Execute(w, map[string]interface{}{
		"book":           book,
		csrf.TemplateTag: b.csrfTemplateTag,
		"session":        b.session,
		"username":       username,
	})
}

func RenderBookConfirmDelete(w io.Writer, b baseViewData, book *data.Book, username string) error {
	return bookConfirmDelete.Execute(w, map[string]interface{}{
		"book":           book,
		csrf.TemplateTag: b.csrfTemplateTag,
		"session":        b.session,
		"username":       username,
	})
}

func RenderBookNew(w io.Writer, b baseViewData, form BookEditForm, verr validate.Errors, username string) error {
	return bookNew.Execute(w, map[string]interface{}{
		"fields":         form,
		"errors":         verr,
		csrf.TemplateTag: b.csrfTemplateTag,
		"session":        b.session,
		"username":       username,
	})
}

func RenderBookImportCSVForm(w io.Writer, b baseViewData, username string) error {
	return bookImportCSVForm.Execute(w, map[string]interface{}{
		csrf.TemplateTag: b.csrfTemplateTag,
		"session":        b.session,
		"username":       username,
	})
}
