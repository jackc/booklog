package server

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/domain"
)

func UserRegistrationNew(w http.ResponseWriter, r *http.Request) {
	var rua domain.RegisterUserArgs

	err := RenderUserRegistrationNew(w, csrf.TemplateField(r), rua, nil)
	if err != nil {
		panic(err)
	}
}
