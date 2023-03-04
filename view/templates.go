package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"

	"github.com/jackc/booklog/route"
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

func RootTemplate() *template.Template {
	fsys := os.DirFS(os.Getenv("TEMP_TEMPLATE_DIR"))

	var err error
	rootTmpl, err := loadTemplates(fsys)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return rootTmpl
}
