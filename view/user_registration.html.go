package view

import (
	"html"
	"io"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
)

func UserRegistration(w io.Writer, bva *BaseViewArgs, form data.RegisterUserArgs, verr validate.Errors) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
  <header>Sign Up</header>

  <form action="`)
	io.WriteString(w, html.EscapeString(route.UserRegistrationPath()))
	io.WriteString(w, `" method="post">
    `)
	io.WriteString(w, bva.CSRFField)
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

    <button type="submit">Sign up</button>
    <a href="`)
	io.WriteString(w, html.EscapeString(route.NewLoginPath()))
	io.WriteString(w, `">Login</a>
  </form>
</div>
`)

	return nil
}
