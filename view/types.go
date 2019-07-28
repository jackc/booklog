package view

import "github.com/jackc/booklog/data"

type BaseViewArgs struct {
	CSRFField   string
	CurrentUser *data.UserMin
	PathUser    *data.UserMin
}
