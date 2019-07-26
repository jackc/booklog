package server

import (
	"fmt"
	"html/template"
)

var RouteFuncMap = template.FuncMap{
	"NewBookPath":             NewBookPath,
	"BookPath":                BookPath,
	"BookConfirmDeletePath":   BookConfirmDeletePath,
	"EditBookPath":            EditBookPath,
	"BooksPath":               BooksPath,
	"NewUserRegistrationPath": NewUserRegistrationPath,
	"UserRegistrationPath":    UserRegistrationPath,
	"NewLoginPath":            NewLoginPath,
	"LoginPath":               LoginPath,
	"LogoutPath":              LogoutPath,
	"ImportBookCSVFormPath":   ImportBookCSVFormPath,
	"ImportBookCSVPath":       ImportBookCSVPath,
	"ExportBookCSVPath":       ExportBookCSVPath,
}

func BooksPath(username string) string {
	return fmt.Sprintf("/users/%s/books", username)
}

func BookPath(username string, id int64) string {
	return fmt.Sprintf("/users/%s/books/%d", username, id)
}

func BookConfirmDeletePath(username string, id int64) string {
	return fmt.Sprintf("/users/%s/books/%d/confirm_delete", username, id)
}

func EditBookPath(username string, id int64) string {
	return fmt.Sprintf("/users/%s/books/%d/edit", username, id)
}

func NewBookPath(username string) string {
	return fmt.Sprintf("/users/%s/books/new", username)
}

func ImportBookCSVFormPath(username string) string {
	return fmt.Sprintf("/users/%s/books/import_csv/form", username)
}

func ImportBookCSVPath(username string) string {
	return fmt.Sprintf("/users/%s/books/import_csv", username)
}

func ExportBookCSVPath(username string) string {
	return fmt.Sprintf("/users/%s/books.csv", username)
}

func NewUserRegistrationPath() string {
	return "/user_registration/new"
}

func UserRegistrationPath() string {
	return "/user_registration"
}

func NewLoginPath() string {
	return "/login"
}

func LoginPath() string {
	return "/login/handle"
}

func LogoutPath() string {
	return "/logout"
}
