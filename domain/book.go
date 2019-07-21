package domain

import (
	"context"
	"errors"

	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type queryExecer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

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

	_, err := db.Exec(ctx, "insert into finished_book(reader_id, title, author, date_finished, media) values($1, $2, $3, $4, $5)",
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
	IDString     string
	ID           int64
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func UpdateBook(ctx context.Context, db queryExecer, args UpdateBookArgs) error {
	v := validate.New()
	v.Presence("title", args.Title)
	v.Presence("author", args.Author)
	v.Presence("dateFinished", args.DateFinished)
	v.Presence("media", args.Media)

	if v.Err() != nil {
		return v.Err()
	}

	commandTag, err := db.Exec(ctx, "update finished_book set title=$1, author=$2, date_finished=$3, media=$4 where id=$5",
		args.Title,
		args.Author,
		args.DateFinished,
		args.Media,
		args.ID)
	if err != nil {
		return err
	}
	if string(commandTag) != "UPDATE 1" {
		return errors.New("not found")
	}

	return nil
}

type DeleteBookArgs struct {
	IDString string
	ID       int64
}

func DeleteBook(ctx context.Context, db queryExecer, args DeleteBookArgs) error {
	_, err := db.Exec(ctx, "delete from finished_book where id=$1", args.ID)
	return err
}
