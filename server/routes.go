package server

import (
	"fmt"
	"html/template"
)

var RouteFuncMap = template.FuncMap{
	"newBookPath":          NewBookPath,
	"bookPath":             BookPath,
	"editBookPath":         EditBookPath,
	"booksPath":            BooksPath,
	"userRegistrationPath": UserRegistrationPath,
	"newLoginPath":         NewLoginPath,
	"loginPath":            LoginPath,
	"logoutPath":           LogoutPath,
}

func BooksPath(username string) string {
	return fmt.Sprintf("/users/%s/books", username)
}

func BookPath(username string, id int64) string {
	return fmt.Sprintf("/users/%s/books/%d", username, id)
}

func EditBookPath(username string, id int64) string {
	return fmt.Sprintf("/users/%s/books/%d/edit", username, id)
}

func NewBookPath(username string) string {
	return fmt.Sprintf("/users/%s/books/new", username)
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
