package view

import "github.com/jackc/booklog/data"

type BaseViewArgs struct {
	CSRFField   string
	CurrentUser *data.UserMin
	PathUser    *data.UserMin
}

type YearBookList struct {
	Year  int
	Books []*data.Book
}
