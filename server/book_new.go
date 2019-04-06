package server

import (
	"html/template"
	"net/http"
)

type BookNew struct {
}

func (action *BookNew) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO - CSRF protection

	tmpl := template.New("new")
	tmpl.Funcs(template.FuncMap{"createBookPath": CreateBookPath})

	tmpl, err := tmpl.Parse(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>Books I Have Read</title>
  </head>
  <body>
    <form action="{{ createBookPath }}" method="post">
      <div>
        <label for="title">Title</label>
        <input type="text" name="title" id="title">
      </div>

      <div>
        <label for="author">Author</label>
        <input type="text" name="author" id="author">
      </div>

      <div>
        <label for="dateFinished">Date Finished</label>
        <input type="date" name="dateFinished" id="dateFinished">
      </div>

      <div>
        <label for="media">Media</label>
        <input type="text" name="media" id="media">
      </div>

      <button type="submit">Save</button>

    </form>
  </body>
</html>`)

	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}
