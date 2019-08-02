package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
)

func BookConfirmDelete(w io.Writer, bva *BaseViewArgs, book *data.Book) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
  <h2>Confirm you want to delete this book?</h2>
  <dl>
    <dt>Title</dt>
    <dd>`)
	io.WriteString(w, html.EscapeString(book.Title))
	io.WriteString(w, `</dd>
    <dt>Author</dt>
    <dd>`)
	io.WriteString(w, html.EscapeString(book.Author))
	io.WriteString(w, `</dd>
    <dt>Finish Date</dt>
    <dd>`)
	io.WriteString(w, html.EscapeString(book.FinishDate.Format("January 2, 2006")))
	io.WriteString(w, `</dd>
    <dt>Media</dt>
    <dd>`)
	io.WriteString(w, html.EscapeString(book.Media))
	io.WriteString(w, `</dd>
  </dl>

  <form action="`)
	io.WriteString(w, html.EscapeString(route.BookPath(bva.PathUser.Username, book.ID)))
	io.WriteString(w, `" method="post">
    <input type="hidden" name="_method" value="DELETE">
    `)
	io.WriteString(w, bva.CSRFField)
	io.WriteString(w, `
    <button type="submit" class="btn">Delete</button>
  </form>
`)
	LayoutFooter(w, bva)
	io.WriteString(w, `
`)

	return nil
}
