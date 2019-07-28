package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/route"
)

func LayoutHeader(w io.Writer, bva *BaseViewArgs) error {
	io.WriteString(w, `<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Booklog</title>
    <link rel="stylesheet" href="/static/css/main.css">
  </head>
  <body>
    <header>
      <h1><a href="/">Booklog</a></h1>
      <nav>
        <ul>
          `)
	if bva.PathUser != nil {
		io.WriteString(w, `
            <li><a href="`)
		io.WriteString(w, html.EscapeString(route.NewBookPath(bva.PathUser.Username)))
		io.WriteString(w, `">New Book</a></li>
            <li><a href="`)
		io.WriteString(w, html.EscapeString(route.ImportBookCSVFormPath(bva.PathUser.Username)))
		io.WriteString(w, `">Import</a></li>
            <li><a href="`)
		io.WriteString(w, html.EscapeString(route.ExportBookCSVPath(bva.PathUser.Username)))
		io.WriteString(w, `">Export</a></li>
          `)
	}
	io.WriteString(w, `
          `)
	if bva.CurrentUser != nil {
		io.WriteString(w, `
            <li>
              <form action="`)
		io.WriteString(w, html.EscapeString(route.LogoutPath()))
		io.WriteString(w, `" method="POST">
                `)
		io.WriteString(w, bva.CSRFField)
		io.WriteString(w, `
                <button>Logout</button>
              </form>
            </li>
          `)
	} else {
		io.WriteString(w, `
            <li><a href="`)
		io.WriteString(w, html.EscapeString(route.NewLoginPath()))
		io.WriteString(w, `">Login</a></li>
          `)
	}
	io.WriteString(w, `
        </ul>
      </nav>
    </header>
    <div class="content">
`)

	return nil
}
