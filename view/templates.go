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
	cache *cachet.Cache[*template.Template]
}

func NewHTMLTemplateRenderer(templatePath string, liveReload bool) *HTMLTemplateRenderer {
	cache := &cachet.Cache[*template.Template]{
		Load: func() (*template.Template, error) {
			fsys := os.DirFS(templatePath)
			return loadTemplates(fsys)
		},
	}

	if liveReload {
		// Reload every time. We could track the modified timestamp of all files and only reload when needed but that would
		// still require walking templatePath for each request. It would be faster than reloading all templates every time
		// but it has not been an issue at this point. Beyond that, fsnotify could be used, but that would bring in an
		// external dependency and can require changing the max open files.
		//
		// Until performance becomes an issue do the simplest thing that can possibly work.
		cache.IsStale = func() (bool, error) {
			return true, nil
		}
	}

	return &HTMLTemplateRenderer{
		cache: cache,
	}
}

func (htr *HTMLTemplateRenderer) ExecuteTemplate(wr io.Writer, name string, data any) error {
	rootTemplate, err := htr.cache.Get()
	if err != nil {
		return fmt.Errorf("failed to get root template from cache: %w", err)
	}

	return rootTemplate.ExecuteTemplate(wr, name, data)
}
