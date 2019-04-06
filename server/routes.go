package server

import "fmt"

func NewBookPath() string {
	return "/new"
}

func CreateBookPath() string {
	return "/books"
}

func DeleteBookPath(id string) string {
	return fmt.Sprintf("/books/%s/delete", id)
}
