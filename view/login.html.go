package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
)

func Login(w io.Writer, bva *BaseViewArgs, form data.UserLoginArgs, verr validate.Errors) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
  <header>Login</header>

  <form action="`)
	io.WriteString(w, html.EscapeString(route.LoginPath()))
	io.WriteString(w, `" method="post">
    `)
	io.WriteString(w, bva.CSRFField)
	io.WriteString(w, `

    `)
	if errs, ok := verr["base"]; ok {
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

    <div class="field">
      <label for="username">Username</label>
      <input type="text" name="username" id="username" value="`)
	io.WriteString(w, html.EscapeString(form.Username))
	io.WriteString(w, `" autofocus required>
      `)
	if errs, ok := verr["username"]; ok {
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
      <label for="password">Password</label>
      <input type="password" name="password" id="password" value="`)
	io.WriteString(w, html.EscapeString(form.Password))
	io.WriteString(w, `" required minlength="8">
      `)
	if errs, ok := verr["password"]; ok {
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

    <button type="submit" class="btn">Login</button>
    <a href="`)
	io.WriteString(w, html.EscapeString(route.NewUserRegistrationPath()))
	io.WriteString(w, `">Sign up</a>
  </form>
</div>
`)
	LayoutFooter(w, bva)
	io.WriteString(w, `
`)

	return nil
}
