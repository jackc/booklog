package server

import (
	"net/http"

	"github.com/gorilla/csrf"
)

type UserRegistrationRequest struct {
	Username string
	Password string
}

func UserRegistrationNew(w http.ResponseWriter, r *http.Request) {
	urr := &UserRegistrationRequest{}

	err := RenderUserRegistrationNew(w, csrf.TemplateField(r), urr, map[string]string{})
	if err != nil {
		panic(err)
	}
}
