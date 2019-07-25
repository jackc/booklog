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

type CreateBookArgs struct {
	ReaderIDString string
	ReaderID       int64
	Title          string
	Author         string
	DateFinished   string
	Media          string
}

func CreateBook(ctx context.Context, db queryExecer, args CreateBookArgs) error {
	v := validate.New()
	v.Presence("title", args.Title)
	v.Presence("author", args.Author)
	v.Presence("dateFinished", args.DateFinished)
	v.Presence("media", args.Media)

	if v.Err() != nil {
		return v.Err()
	}

	_, err := db.Exec(ctx, "insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5)",
		args.ReaderID,
		args.Title,
		args.Author,
		args.DateFinished,
		args.Media)
	if err != nil {
		return err
	}

	return nil
}

type UpdateBookArgs struct {
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func UpdateBook(ctx context.Context, db queryExecer, currentUserID int64, bookID int64, args UpdateBookArgs) error {
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
	v.Presence("title", args.Title)
	v.Presence("author", args.Author)
	v.Presence("dateFinished", args.DateFinished)
	v.Presence("media", args.Media)

	if v.Err() != nil {
		return v.Err()
	}

	commandTag, err := db.Exec(ctx, "update books set title=$1, author=$2, finish_date=$3, media=$4 where id=$5",
		args.Title,
		args.Author,
		args.DateFinished,
		args.Media,
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
