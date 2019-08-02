package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
)

func BookEdit(w io.Writer, bva *BaseViewArgs, bookID int64, form BookEditForm, verr validate.Errors) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
    <header>New Book</header>

    <form action="`)
	io.WriteString(w, html.EscapeString(route.BookPath(bva.PathUser.Username, bookID)))
	io.WriteString(w, `" method="post">
      <input type="hidden" name="_method" value="PATCH">
      `)
	BookFormFields(w, bva, form, verr)
	io.WriteString(w, `
    </form>
  </div>
`)
	LayoutFooter(w, bva)
	io.WriteString(w, `
`)

	return nil
}
