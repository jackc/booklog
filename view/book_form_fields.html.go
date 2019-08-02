package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/validate"
)

func BookFormFields(w io.Writer, bva *BaseViewArgs, form BookEditForm, verr validate.Errors) error {
	io.WriteString(w, bva.CSRFField)
	io.WriteString(w, `

<div class="field">
  <label for="title">Title</label>
  <input type="text" name="title" id="title" value="`)
	io.WriteString(w, html.EscapeString(form.Title))
	io.WriteString(w, `" >
  `)
	if errs, ok := verr["title"]; ok {
		io.WriteString(w, `
    `)
		for _, e := range errs {
			io.WriteString(w, `
      <div class="error">`)
			io.WriteString(w, html.EscapeString(e.Error()))
			io.WriteString(w, `</div>
    `)
		}
		io.WriteString(w, `
  `)
	}
	io.WriteString(w, `
</div>

<div class="field">
  <label for="author">Author</label>
  <input type="text" name="author" id="author" value="`)
	io.WriteString(w, html.EscapeString(form.Author))
	io.WriteString(w, `" >
  `)
	if errs, ok := verr["author"]; ok {
		io.WriteString(w, `
    `)
		for _, e := range errs {
			io.WriteString(w, `
      <div class="error">`)
			io.WriteString(w, html.EscapeString(e.Error()))
			io.WriteString(w, `</div>
    `)
		}
		io.WriteString(w, `
  `)
	}
	io.WriteString(w, `
</div>

<div class="field">
  <label for="finishDate">Finish Date</label>
  <input type="date" name="finishDate" id="finishDate" value="`)
	io.WriteString(w, html.EscapeString(form.FinishDate))
	io.WriteString(w, `" >
  `)
	if errs, ok := verr["finishDate"]; ok {
		io.WriteString(w, `
    `)
		for _, e := range errs {
			io.WriteString(w, `
      <div class="error">`)
			io.WriteString(w, html.EscapeString(e.Error()))
			io.WriteString(w, `</div>
    `)
		}
		io.WriteString(w, `
  `)
	}
	io.WriteString(w, `
</div>

<div class="field">
  <label for="media">Media</label>
  <select name="media" id="media">
    <option `)
	if form.Media == "book" {
		io.WriteString(w, `selected`)
	}
	io.WriteString(w, `>book</option>
    <option `)
	if form.Media == "audiobook" {
		io.WriteString(w, `selected`)
	}
	io.WriteString(w, `>audiobook</option>
    <option `)
	if form.Media == "video" {
		io.WriteString(w, `selected`)
	}
	io.WriteString(w, `>video</option>
  </select>
  `)
	if errs, ok := verr["media"]; ok {
		io.WriteString(w, `
    `)
		for _, e := range errs {
			io.WriteString(w, `
      <div class="error">`)
			io.WriteString(w, html.EscapeString(e.Error()))
			io.WriteString(w, `</div>
    `)
		}
		io.WriteString(w, `
  `)
	}
	io.WriteString(w, `
</div>

<button type="submit" class="btn">Save</button>
`)

	return nil
}
