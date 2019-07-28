package view

import "io"

func LayoutFooter(w io.Writer, bva *BaseViewArgs) error {
	io.WriteString(w, `</div>
</body>
</html>
`)

	return nil
}
