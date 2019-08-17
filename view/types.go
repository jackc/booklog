package view

import (
	"errors"
	"time"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/validate"
)

type BaseViewArgs struct {
	CSRFField   string
	CurrentUser *data.UserMin
	PathUser    *data.UserMin
}

type YearBookList struct {
	Year  int
	Books []*data.Book
}

type BookEditForm struct {
	Title      string
	Author     string
	FinishDate string
	Format     string
}

func (f BookEditForm) Parse() (data.Book, validate.Errors) {
	var err error
	book := data.Book{
		Title:  f.Title,
		Author: f.Author,
		Format: f.Format,
	}
	v := validate.New()

	dateFormats := []string{"2006-01-02", "1/2/2006", "1/2/06"}

	for _, df := range dateFormats {
		book.FinishDate, err = time.Parse(df, f.FinishDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		v.Add("finishDate", errors.New("is not a date"))
	}

	if v.Err() != nil {
		return book, v.Err().(validate.Errors)
	}

	return book, nil
}
