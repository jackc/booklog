package view

import "io"

func BookImportCSVForm(w io.Writer, bva *BaseViewArgs) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<div class="card">
  <header>Import Book CSV</header>

  <p>CSV must include header row.</p>
  <p>CSV must include 4 columns in order: title, author, date finished, and media.</p>

  <form enctype="multipart/form-data" action="{{ route.ImportBookCSVPath(bva.PathUser.Username) %>" method="post">
    `)
	io.WriteString(w, bva.CSRFField)
	io.WriteString(w, `

    <div class="field">
      <label for="file">Media</label>
      <input type="file" name="file" id="file" />
    </div>

    <button type="submit">Import</button>
  </form>
</div>
`)
	LayoutFooter(w, bva)
	io.WriteString(w, `
`)

	return nil
}