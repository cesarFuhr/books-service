package pkgerrors

type ErrResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (e ErrResponse) Error() string {
	return e.Message
}

var ErrResponseBookEntryBlankFileds = ErrResponse{100, "all the fields - name, price and inventory - must be filled correctly."}
var ErrResponseBookNotFound = ErrResponse{101, "book not found"}
var ErrResponseBookEntryInvalidJSON = ErrResponse{102, "invalid json request."}
var ErrResponseIdInvalidFormat = ErrResponse{103, "the endpoint is not a valid format ID. Must be /books/{uuid}"}
var ErrResponseQueryPriceInvalidFormat = ErrResponse{104, "query parameter 'price' must be a float between 0 and 9999.99"}
var ErrResponseQuerySortByInvalid = ErrResponse{105, "query parameter 'sort_by' must be: name, price, inventory, created_at or updated_at. 'sort_direction' must be asc or desc."}
var ErrResponseQueryPageInvalid = ErrResponse{106, "query parameter 'page' must be an int starting in 1. 'page_size' must be an int beetween 1 and 30."}
var ErrResponseQueryPageOutOfRange = ErrResponse{107, "page out of range."}
