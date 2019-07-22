package server

import (
	"net/http"

	"github.com/jackc/booklog/domain"
)

func UserRegistrationNew(w http.ResponseWriter, r *http.Request) {
	var rua domain.RegisterUserArgs

	err := RenderUserRegistrationNew(w, baseViewDataFromRequest(r), rua, nil)
	if err != nil {
		panic(err)
	}
}
