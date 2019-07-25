package domain

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
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

func CreateBook(ctx context.Context, db queryExecer, currentUserID int64, ownerID int64, attrs BookAttrs) error {
	if ownerID != currentUserID {
		return &ForbiddenError{currentUserID: currentUserID, msg: fmt.Sprintf("create book for user_id=%d", ownerID)}
	}

	v := validate.New()
	v.Presence("title", attrs.Title)
	v.Presence("author", attrs.Author)
	v.Presence("media", attrs.Media)

	if v.Err() != nil {
		return v.Err()
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

	v := validate.New()
	v.Presence("title", attrs.Title)
	v.Presence("author", attrs.Author)
	v.Presence("media", attrs.Media)

	if v.Err() != nil {
		return v.Err()
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

// TODO - need DB transaction control - so queryExecer is insufficient
func ImportBooksFromCSV(ctx context.Context, db queryExecer, userID int64, r io.Reader) error {
	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return errors.New("CSV must have at least 2 rows")
	}

	if len(records[0]) < 4 {
		return errors.New("CSV must have at least 4 columns")
	}

	for i, record := range records[1:] {
		v := validate.New()
		v.Presence("title", record[0])
		v.Presence("author", record[1])
		v.Presence("dateFinished", record[2])
		if record[3] == "" {
			record[3] = "book"
		}
		v.Presence("media", record[3])

		if v.Err() != nil {
			return v.Err()
		}

		var dateFinished time.Time
		dateFinished, err = time.Parse("1/2/2006", record[2])
		if err != nil {
			return errors.Errorf("row %d: %v", i, err)
		}

		_, err := db.Exec(ctx, "insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5)",
			userID,
			record[0],
			record[1],
			dateFinished,
			record[3])
		if err != nil {
			return err
		}
	}

	return nil
}
