package data

import (
	"context"
)

type BooksPerYearItem struct {
	Count int32
	Year  int16
}

func BooksPerYear(ctx context.Context, db queryExecer, userID int64) ([]BooksPerYearItem, error) {
	var booksPerYear []BooksPerYearItem

	rows, err := db.Query(ctx, "select date_part('year', finish_date), count(*) from books where user_id=$1 group by 1 order by 1 desc", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item BooksPerYearItem
		rows.Scan(&item.Year, &item.Count)
		booksPerYear = append(booksPerYear, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return booksPerYear, nil
}
