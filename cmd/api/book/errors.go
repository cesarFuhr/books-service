package book

import (
	"fmt"
)

type ErrResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (e ErrResponse) Error() string {
	return e.Message
}

var ErrResponseBookEntryBlankFields = ErrResponse{100, "all the fields - name, price and inventory - must be filled correctly."}
var ErrResponseBookNotFound = ErrResponse{101, "book not found"}
var ErrResponseEntryInvalidJSON = ErrResponse{102, "invalid json request."}
var ErrResponseIdInvalidFormat = ErrResponse{103, "the endpoint is not a valid format ID. Must be /books/{uuid}"}
var ErrResponseQueryPriceInvalidFormat = ErrResponse{104, "query parameter 'price' must be a float between 0 and 9999.99"}
var ErrResponseQuerySortByInvalid = ErrResponse{105, "query parameter 'sort_by' must be: name, price, inventory, created_at or updated_at. 'sort_direction' must be asc or desc."}
var ErrResponseQueryPageInvalid = ErrResponse{106, "query parameter 'page' must be an int starting in 1. 'page_size' must be an int beetween 1 and 30."}
var ErrResponseQueryPageOutOfRange = ErrResponse{107, "page out of range."}
var ErrResponseRequestTimeout = ErrResponse{109, "context deadline exceeded"}
var ErrResponseOrderNotFound = ErrResponse{110, "order not found"}
var ErrResponseOrderNotAcceptingItems = ErrResponse{111, "order not accepting items"}
var ErrResponseBookIsArchived = ErrResponse{112, "book status is archived"}
var ErrResponseInsufficientInventory = ErrResponse{113, "inventory is insufficient for this order"}
var ErrResponseBookNotAtOrder = ErrResponse{114, "book is not at the order"}
var ErrResponseUpdateOrderEntryBlankFields = ErrResponse{116, "all the fields - order_id, book_id and book_units_to_add - must be filled correctly."}
var ErrResponseNewOrderEntryBlankFields = ErrResponse{117, "field user_id must be filled correctly."}
var ErrResponseListOrderItemsEntryBlankFields = ErrResponse{118, "field order_id must be filled correctly."}

type ErrNotificationFailed struct {
	statusCode int
}

func (e ErrNotificationFailed) Error() string {
	return fmt.Sprintf("ntfy wrong response - want: 200 OK, got: %d", e.statusCode)
}

func NewErrNotificationFailed(statusCode int) ErrNotificationFailed {
	return ErrNotificationFailed{statusCode: statusCode}
}
