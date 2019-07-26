package domain

import (
	"context"
	"fmt"
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

func CreateBook(ctx context.Context, db queryExecer, currentUserID int64, ownerID int64, attrs BookAttrs) error {
	if ownerID != currentUserID {
		return &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("create book for user_id=%d", ownerID)}
	}

	if verrs := attrs.Validate(); verrs != nil {
		return verrs
	}

	_, err := db.Exec(ctx, "insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5)",
		ownerID,
		attrs.Title,
		attrs.Author,
		attrs.DateFinished,
		attrs.Media)
	if err != nil {
		return err
	}

	return nil
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
