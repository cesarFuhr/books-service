package book

import (
	"math"
	"net/url"
	"strconv"
	"time"

	bookerrors "github.com/books-service/cmd/api/errors"
	"github.com/google/uuid"
)

type Book struct { //MOVE JSON TAGS TO HTTP PACKAGE!!
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     *float32  `json:"price"`
	Inventory *int      `json:"inventory"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Archived  bool      `json:"archived"`
}

/* Verifies if all entry fields are filled and returns a warning message if so. */
func FilledFields(bookEntry Book) error {
	if bookEntry.Name == "" {
		return bookerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Price == nil {
		return bookerrors.ErrResponseBookEntryBlankFileds
	}
	if bookEntry.Inventory == nil {
		return bookerrors.ErrResponseBookEntryBlankFileds
	}

	return nil
}

/*Validates and prepares the ordering parameters of the query.*/
func extractOrderParams(query url.Values) (sortBy string, sortDirection string, valid bool) {
	sortDirection = query.Get("sort_direction")
	switch sortDirection {
	case "":
		sortDirection = "asc"
	case "asc":
		break
	case "desc":
		break
	default:
		return sortBy, sortDirection, false
	}

	sortBy = query.Get("sort_by")
	switch sortBy {
	case "":
		sortBy = "name"
	case "name":
		break
	case "price":
		break
	case "inventory":
		break
	case "created_at":
		break
	case "updated_at":
		break
	default:
		return sortBy, sortDirection, false
	}

	return sortBy, sortDirection, true
}

/*Validates and prepares the pagination parameters of the query.*/
func pagination(query url.Values, itemsTotal int) (pagesTotal, page, pageSize int, err error) {

	pageStr := query.Get("page") //Convert page value to int and set default to 1.
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
		if page <= 0 {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
	}

	pageSizeStr := query.Get("page_size") //Convert page_size value to int and set default to 10.
	if pageSizeStr == "" {
		pageSize = 10
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
		if !(0 < pageSize && pageSize < 31) {
			return 0, 0, 0, bookerrors.ErrResponseQueryPageInvalid
		}
	}

	pagesTotal = int(math.Ceil(float64(itemsTotal) / float64(pageSize)))
	if page > pagesTotal {
		return 0, 0, 0, bookerrors.ErrResponseQueryPageOutOfRange
	}

	return pagesTotal, page, pageSize, nil
}
