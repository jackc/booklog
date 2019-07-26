package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

type BookAttrs struct {
	Title        string
	Author       string
	DateFinished time.Time
	Media        string
}

func (attrs *BookAttrs) Normalize() {
	attrs.Title = strings.TrimSpace(attrs.Title)
	attrs.Author = strings.TrimSpace(attrs.Author)
	attrs.Media = strings.TrimSpace(attrs.Media)
}

func (attrs BookAttrs) Validate() validate.Errors {
	v := validate.New()
	v.Presence("title", attrs.Title)
	v.Presence("author", attrs.Author)
	v.Presence("media", attrs.Media)

	if v.Err() != nil {
		return v.Err().(validate.Errors)
	}

	return nil
}

func CreateBook(ctx context.Context, db queryExecer, currentUserID int64, ownerID int64, attrs BookAttrs) (int64, error) {
	if ownerID != currentUserID {
		return 0, &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("create book for user_id=%d", ownerID)}
	}

	attrs.Normalize()
	if verrs := attrs.Validate(); verrs != nil {
		return 0, verrs
	}

	var bookID int64
	err := db.QueryRow(ctx, "insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id",
		ownerID,
		attrs.Title,
		attrs.Author,
		attrs.DateFinished,
		attrs.Media,
	).Scan(&bookID)
	if err != nil {
		return 0, err
	}

	return bookID, nil
}

func UpdateBook(ctx context.Context, db queryExecer, currentUserID int64, bookID int64, attrs BookAttrs) error {
	var ownerID int64
	err := db.QueryRow(ctx, "select user_id from books where id=$1", bookID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
		} else {
			return err
		}
	}

	if ownerID != currentUserID {
		return &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("delete book id=%d", bookID)}
	}

	attrs.Normalize()
	if verrs := attrs.Validate(); verrs != nil {
		return verrs
	}

	commandTag, err := db.Exec(ctx, "update books set title=$1, author=$2, finish_date=$3, media=$4 where id=$5",
		attrs.Title,
		attrs.Author,
		attrs.DateFinished,
		attrs.Media,
		bookID)
	if err != nil {
		return err
	}
	if string(commandTag) != "UPDATE 1" {
		return errors.New("not found")
	}

	return nil
}

// DeleteBook deletes the book specified by bookID at the behest of currentUserID. It returns a NotFoundError if the book
// cannot be found and a ForbiddenError is the user does not have permission to delete the book.
func DeleteBook(ctx context.Context, db queryExecer, currentUserID int64, bookID int64) error {
	// TODO - these two queries should run in a transaction
	var ownerID int64
	err := db.QueryRow(ctx, "select user_id from books where id=$1", bookID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
		} else {
			return err
		}
	}

	if ownerID != currentUserID {
		return &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("delete book id=%d", bookID)}
	}

	_, err = db.Exec(ctx, "delete from books where id=$1", bookID)
	return err
}

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

func GetBook(ctx context.Context, db queryExecer, currentUserID int64, bookID int64) (*Book, error) {
	var book Book
	err := db.QueryRow(ctx, "select id, user_id, title, author, finish_date, media, insert_time, update_time from books where id=$1", bookID).
		Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.FinishDate, &book.Media, &book.InsertTime, &book.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{target: fmt.Sprintf("book id=%d", bookID)}
		}
		return nil, err
	}

	if book.UserID != currentUserID {
		return nil, &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("select book id=%d", bookID)}
	}

	return &book, nil
}
