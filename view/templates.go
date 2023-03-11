package view

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"

	"github.com/jackc/booklog/route"
	"github.com/jackc/cachet"
)

func LoadManifest(path string) (map[string]string, error) {
	manifestBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("LoadManifest: %w", err)
	}

	var manifest map[string]any
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return nil, fmt.Errorf("LoadManifest %s: %w", path, err)
	}

	assetMap := make(map[string]string, len(manifest))
	for k, v := range manifest {
		assetMap["/"+k] = "/" + v.(map[string]any)["file"].(string)
	}

	return assetMap, nil
}

func loadTemplates(fsys fs.FS, rootTmpl *template.Template) error {
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
		return err
	}

	return nil
}

type HTMLTemplateRenderer struct {
	cache *cachet.Cache[*template.Template]
}

func NewHTMLTemplateRenderer(templatePath string, assetMap map[string]string, liveReload bool) *HTMLTemplateRenderer {
	funcMap := template.FuncMap{
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
	}

	if assetMap == nil {
		funcMap["assetPath"] = func(name string) (string, error) {
			return name, nil
		}
	} else {
		funcMap["assetPath"] = func(name string) (string, error) {
			path, ok := assetMap[name]
			if !ok {
				return "", fmt.Errorf("unknown asset: %v", name)
			}

			return path, nil
		}
	}
	cache := &cachet.Cache[*template.Template]{
		Load: func() (*template.Template, error) {
			fsys := os.DirFS(templatePath)
			tmpl := template.New("root").Funcs(funcMap)
			err := loadTemplates(fsys, tmpl)
			if err != nil {
				return nil, err
			}

			return tmpl, nil
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
