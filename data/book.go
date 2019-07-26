package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

type Book struct {
	ID         int64
	UserID     int64
	Title      string
	Author     string
	FinishDate time.Time
	Media      string
	InsertTime time.Time
	UpdateTime time.Time
}

func (book *Book) Normalize() {
	book.Title = strings.TrimSpace(book.Title)
	book.Author = strings.TrimSpace(book.Author)
	book.Media = strings.TrimSpace(book.Media)
}

func (book *Book) Validate() validate.Errors {
	v := validate.New()
	v.Presence("title", book.Title)
	v.Presence("author", book.Author)
	v.Presence("media", book.Media)

	if v.Err() != nil {
		return v.Err().(validate.Errors)
	}

	return nil
}

// CreateBook inserts a book into the database. It ignores the ID, InsertTime, and UpdateTime fields.
func CreateBook(ctx context.Context, db queryExecer, book Book) (*Book, error) {
	book.Normalize()
	if verrs := book.Validate(); verrs != nil {
		return nil, verrs
	}

	err := db.QueryRow(ctx, "insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id, insert_time, update_time",
		book.UserID,
		book.Title,
		book.Author,
		book.FinishDate,
		book.Media,
	).Scan(&book.ID, &book.InsertTime, &book.UpdateTime)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

// Update book updates the Title, Author, FinishDate, and Media fields of book in the database. It uses book.ID as the
// row ID to update.
func UpdateBook(ctx context.Context, db queryExecer, book Book) error {
	book.Normalize()
	if verrs := book.Validate(); verrs != nil {
		return verrs
	}

	commandTag, err := db.Exec(ctx, "update books set title=$1, author=$2, finish_date=$3, media=$4 where id=$5",
		book.Title,
		book.Author,
		book.FinishDate,
		book.Media,
		book.ID)
	if err != nil {
		return err
	}
	if string(commandTag) != "UPDATE 1" {
		return &NotFoundError{target: fmt.Sprintf("book id=%d", book.ID)}
	}

	return nil
}

// DeleteBook deletes the book specified by bookID. It returns a NotFoundError if the book
// cannot be found.
func DeleteBook(ctx context.Context, db queryExecer, bookID int64) error {
	commandTag, err := db.Exec(ctx, "delete from books where id=$1", bookID)
	if string(commandTag) != "DELETE 1" {
		return &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
	}
	return err
}

func GetBook(ctx context.Context, db queryExecer, bookID int64) (*Book, error) {
	var book Book
	err := db.QueryRow(ctx, "select id, user_id, title, author, finish_date, media, insert_time, update_time from books where id=$1", bookID).
		Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.FinishDate, &book.Media, &book.InsertTime, &book.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
		}
		return nil, err
	}

	return &book, nil
}
