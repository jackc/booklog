package server

import (
	"context"
	"net/http"

	"github.com/jackc/booklog/domain"
)

func UserRegistrationNew(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	var rua domain.RegisterUserArgs

	err := RenderUserRegistrationNew(w, baseViewDataFromRequest(r), rua, nil)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
}
