package server

import (
	"html/template"

	"github.com/jackc/booklog/route"
)

var RouteFuncMap = template.FuncMap{
	"UserHomePath":            route.UserHomePath,
	"NewBookPath":             route.NewBookPath,
	"BookPath":                route.BookPath,
	"BookConfirmDeletePath":   route.BookConfirmDeletePath,
	"EditBookPath":            route.EditBookPath,
	"BooksPath":               route.BooksPath,
	"NewUserRegistrationPath": route.NewUserRegistrationPath,
	"UserRegistrationPath":    route.UserRegistrationPath,
	"NewLoginPath":            route.NewLoginPath,
	"LoginPath":               route.LoginPath,
	"LogoutPath":              route.LogoutPath,
	"ImportBookCSVFormPath":   route.ImportBookCSVFormPath,
	"ImportBookCSVPath":       route.ImportBookCSVPath,
	"ExportBookCSVPath":       route.ExportBookCSVPath,
}
