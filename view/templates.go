package view

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"

	"github.com/jackc/booklog/route"
	"github.com/jackc/cachet"
)

var cache *cachet.Cache[*template.Template]

func init() {
	cache = &cachet.Cache[*template.Template]{
		Load: func() (*template.Template, error) {
			fsys := os.DirFS(os.Getenv("TEMP_TEMPLATE_DIR"))
			return loadTemplates(fsys)
		},
		IsStale: func() (bool, error) {
			return false, nil
		},
	}
}

func loadTemplates(fsys fs.FS) (*template.Template, error) {
	rootTmpl := template.New("root").Funcs(template.FuncMap{
		"UserHomePath":            route.UserHomePath,
		"BooksPath":               route.BooksPath,
		"BookPath":                route.BookPath,
		"BookConfirmDeletePath":   route.BookConfirmDeletePath,
		"EditBookPath":            route.EditBookPath,
		"NewBookPath":             route.NewBookPath,
		"ImportBookCSVFormPath":   route.ImportBookCSVFormPath,
		"ImportBookCSVPath":       route.ImportBookCSVPath,
		"ExportBookCSVPath":       route.ExportBookCSVPath,
		"NewUserRegistrationPath": route.NewUserRegistrationPath,
		"UserRegistrationPath":    route.UserRegistrationPath,
		"NewLoginPath":            route.NewLoginPath,
		"LoginPath":               route.LoginPath,
		"LogoutPath":              route.LogoutPath,
	})

	walkFunc := func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to walk for %s: %v", path, walkErr)
		}

		if d.Type().IsRegular() {
			tmplSrc, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			tmplName := path
			_, err = rootTmpl.New(tmplName).Parse(string(tmplSrc))
			if err != nil {
				return fmt.Errorf("failed to parse for %s: %v", path, err)
			}
		}

		return nil
	}

	err := fs.WalkDir(fsys, ".", walkFunc)
	if err != nil {
		return nil, err
	}

	return rootTmpl, nil
}

type HTMLTemplateRenderer struct {
}

func (htr *HTMLTemplateRenderer) ExecuteTemplate(wr io.Writer, name string, data any) error {
	rootTemplate, err := cache.Get()
	if err != nil {
		return fmt.Errorf("failed to get root template from cache: %w", err)
	}

	return rootTemplate.ExecuteTemplate(wr, name, data)
}
