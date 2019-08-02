package view

import (
	"html"
	"io"
	"strconv"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
)

func UserHome(
	w io.Writer,
	bva *BaseViewArgs,
	yearBookLists []*YearBookList,
	booksPerYear []data.BooksPerTimeItem,
	booksPerMonthForLastYear []data.BooksPerTimeItem,
) error {
	LayoutHeader(w, bva)
	io.WriteString(w, `
<style>
  ol.years {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  ol.years > li {
    margin-bottom: 2rem;
  }

   ol.years > li > h2 {
    font-size: 2rem;
    color: var(--light-text-color);
  }

  ol.books {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  ol.books > li {
    margin: 1rem 0;
    display: grid;
   }

  ol.books time.finished, ol.books .media, ol.books .author {
    color: var(--light-text-color);
  }

  ol.books > li .title {
    display: block;
    font-weight: bold;
  }

  .stats {
    display: grid;
  }

  .books-per-time h2 {
    margin: 0 0 1rem 0;
  }

  .books-per-time table {
    border-collapse: collapse;
  }

  .books-per-time th {
    font-weight: bold;
    color: var(--light-text-color);
    padding: 2px 0;
    text-align: left;
    min-width: 6rem;
  }

  .books-per-time td {
    color: var(--light-text-color);
    text-align: right;
    padding: 2px 0;
  }

@media (max-width: 32rem) {
  .stats {
    grid-template-columns: 1fr;
  }

  ol.years > li > h2 {
    margin: 0;
  }

  ol.books > li > .what {
    margin-left: 2rem;
  }
}

@media not all and (max-width: 32rem) {
  .stats {
    grid-template-columns: 1fr 1fr;
  }

  ol.books > li {
    display: grid;
    grid-template-columns: auto 1fr;
  }

  ol.years > li > h2 {
    margin: 0 0 0 9rem;
  }

  ol.books time.finished, ol.books .media {
    display: block;
    min-width: 8rem;
    text-align: right;
    margin-right: 1rem;
  }
}
</style>

<div class="stats">
  <div class="card books-per-time">
    <h2>Per Year</h2>

    <table>
      `)
	for _, bpt := range booksPerYear {
		io.WriteString(w, `
        <tr>
          <th>`)
		io.WriteString(w, html.EscapeString(bpt.Time.Format("2006")))
		io.WriteString(w, `</th>
          <td>`)
		io.WriteString(w, strconv.FormatInt(int64(bpt.Count), 10))
		io.WriteString(w, `</td>
        </tr>
      `)
	}
	io.WriteString(w, `
    </table>
  </div>

  <div class="card books-per-time">
    <h2>Last Year Per Month</h2>

    <table>
      `)
	for _, bpt := range booksPerMonthForLastYear {
		io.WriteString(w, `
        <tr>
          <th>`)
		io.WriteString(w, html.EscapeString(bpt.Time.Format("January")))
		io.WriteString(w, `</th>
          <td>`)
		io.WriteString(w, strconv.FormatInt(int64(bpt.Count), 10))
		io.WriteString(w, `</td>
        </tr>
      `)
	}
	io.WriteString(w, `
    </table>
  </div>
</div>

<div class="card">
  `)
	for _, ybl := range yearBookLists {
		io.WriteString(w, `
    <ol class="years">
      <li>
        <h2>`)
		io.WriteString(w, strconv.FormatInt(int64(ybl.Year), 10))
		io.WriteString(w, `</h2>
        <ol class="books">
          `)
		for _, book := range ybl.Books {
			io.WriteString(w, `
            <li>
              <div class="when-and-how">
                <time class="finished"
                  datetime="`)
			io.WriteString(w, html.EscapeString(book.FinishDate.Format("2006-01-02")))
			io.WriteString(w, `"
                  title="`)
			io.WriteString(w, html.EscapeString(book.FinishDate.Format("January 2, 2006")))
			io.WriteString(w, `"
                >
                  `)
			io.WriteString(w, html.EscapeString(book.FinishDate.Format("January 2")))
			io.WriteString(w, `
                </time>
                <span class="media">
                  `)

			var icon string
			switch book.Media {
			case "audiobook":
				icon = "🎧"
			case "book":
				icon = "📖"
			case "video":
				icon = "📺"
			}
			io.WriteString(w, `
                  `)
			io.WriteString(w, html.EscapeString(icon))
			io.WriteString(w, `
                </span>
              </div>
              <div class="what">
                <a class="title" href="`)
			io.WriteString(w, route.BookPath(bva.PathUser.Username, book.ID))
			io.WriteString(w, `">
                  `)
			io.WriteString(w, html.EscapeString(book.Title))
			io.WriteString(w, `
                </a>
                <div class="author">`)
			io.WriteString(w, html.EscapeString(book.Author))
			io.WriteString(w, `</div>
              </div>
            </li>
          `)
		}
		io.WriteString(w, `
        </ol>
      </li>
    </ol>
  `)
	}
	io.WriteString(w, `
</div>
`)
	LayoutFooter(w, bva)
	io.WriteString(w, `
`)

	return nil
}
