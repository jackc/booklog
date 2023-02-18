package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxrecord"
)

type BooksPerTimeItem struct {
	Time  time.Time
	Count int32
}

func BooksPerYear(ctx context.Context, db dbconn, userID int64) ([]BooksPerTimeItem, error) {
	return pgxrecord.Select(
		ctx,
		db,
		"select date_trunc('year', finish_date), count(*) from books where user_id=$1 group by 1 order by 1 desc",
		[]any{userID},
		pgx.RowToStructByPos[BooksPerTimeItem],
	)
}

func BooksPerMonthForLastYear(ctx context.Context, db dbconn, userID int64) ([]BooksPerTimeItem, error) {
	return pgxrecord.Select(
		ctx,
		db,
		`select months, count(books.id)
from generate_series(date_trunc('month', now() - '1 year'::interval), date_trunc('month', now()), '1 month') as months
	left join books on date_trunc('month', finish_date) = months and user_id=$1
group by 1
order by 1 desc`,
		[]any{userID},
		pgx.RowToStructByPos[BooksPerTimeItem],
	)
}
