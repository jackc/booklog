package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/booklog/validate"
	"github.com/jackc/errortree"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
)

type Book struct {
	ID         int64
	UserID     int64
	Title      string
	Author     string
	FinishDate time.Time
	Format     string
	Location   string
	InsertTime time.Time
	UpdateTime time.Time
}

func (book *Book) Normalize() {
	book.Title = strings.TrimSpace(book.Title)
	book.Author = strings.TrimSpace(book.Author)
	book.Format = strings.TrimSpace(book.Format)
	book.Location = strings.TrimSpace(book.Location)
}

func (book *Book) Validate() *errortree.Node {
	v := validate.New()
	v.Presence("title", book.Title)
	v.Presence("author", book.Author)

	allowedFormats := map[string]struct{}{"text": struct{}{}, "audio": struct{}{}, "video": struct{}{}}

	v.Presence("format", book.Format)
	if _, ok := allowedFormats[book.Format]; !ok {
		v.Add("finishDate", errors.New(`must be "text", "audio", or "video"`))
	}

	if book.FinishDate.After(time.Now()) {
		v.Add("finishDate", errors.New("cannot be in future"))
	}

	if v.Err() != nil {
		return v.Err().(*errortree.Node)
	}

	return nil
}

// CreateBook inserts a book into the database. It ignores the ID, InsertTime, and UpdateTime fields.
func CreateBook(ctx context.Context, db dbconn, book Book) (*Book, error) {
	book.Normalize()
	if verrs := book.Validate(); verrs != nil {
		return nil, verrs
	}

	var location *string
	if len(book.Location) > 0 {
		location = &book.Location
	}

	err := db.QueryRow(ctx, "insert into books(user_id, title, author, finish_date, format, location) values($1, $2, $3, $4, $5, $6) returning id, insert_time, update_time",
		book.UserID,
		book.Title,
		book.Author,
		book.FinishDate,
		book.Format,
		location,
	).Scan(&book.ID, &book.InsertTime, &book.UpdateTime)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

// Update book updates the Title, Author, FinishDate, and Format fields of book in the database. It uses book.ID as the
// row ID to update.
func UpdateBook(ctx context.Context, db dbconn, book Book) error {
	book.Normalize()
	if verrs := book.Validate(); verrs != nil {
		return verrs
	}

	var location *string
	if len(book.Location) > 0 {
		location = &book.Location
	}

	commandTag, err := db.Exec(ctx, "update books set title=$1, author=$2, finish_date=$3, format=$4, location=$5 where id=$6",
		book.Title,
		book.Author,
		book.FinishDate,
		book.Format,
		location,
		book.ID)
	if err != nil {
		return err
	}
	if commandTag.String() != "UPDATE 1" {
		return &NotFoundError{target: fmt.Sprintf("book id=%d", book.ID)}
	}

	return nil
}

// DeleteBook deletes the book specified by bookID. It returns a NotFoundError if the book
// cannot be found.
func DeleteBook(ctx context.Context, db dbconn, bookID int64) error {
	commandTag, err := db.Exec(ctx, "delete from books where id=$1", bookID)
	if err != nil {
		return err
	}
	if commandTag.String() != "DELETE 1" {
		return &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
	}
	return nil
}

func GetBook(ctx context.Context, db dbconn, bookID int64) (*Book, error) {
	rows, _ := db.Query(ctx, "select id, user_id, title, author, finish_date, format, location, insert_time, update_time from books where id=$1", bookID)
	book, err := pgx.CollectOneRow(rows, RowToAddrOfBook)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
		}
		return nil, err
	}
	return book, nil
}

func RowToAddrOfBook(row pgx.CollectableRow) (*Book, error) {
	var book Book
	err := row.Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.FinishDate, &book.Format, (*zeronull.Text)(&book.Location), &book.InsertTime, &book.UpdateTime)
	return &book, err
}

func GetAllBooks(ctx context.Context, db dbconn, userID int64) ([]*Book, error) {
	rows, _ := db.Query(ctx, `select id, user_id, title, author, finish_date, format, location, insert_time, update_time
from books
where user_id=$1
order by finish_date desc`,
		userID)
	return pgx.CollectRows(rows, RowToAddrOfBook)
}
