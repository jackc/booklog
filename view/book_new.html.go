package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
)

func BookNew(w io.Writer, bva *BaseViewArgs, form BookEditForm, verr validate.Errors) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
  <header>New Book</header>

  <form action="`)
	io.WriteString(w, html.EscapeString(route.BooksPath(bva.PathUser.Username)))
	io.WriteString(w, `" method="post">
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
