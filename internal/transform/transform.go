package transform

import (
	"go-currently-reading/internal/kavita"
)

func Filter(
	books []kavita.BookSummary,
	predicate func(kavita.BookSummary) bool,
) []kavita.BookSummary {

	result := make([]kavita.BookSummary, 0, len(books))

	for _, book := range books {
		if predicate(book) {
			result = append(result, book)
		}
	}

	return result
}
