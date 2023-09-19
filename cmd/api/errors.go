package main

type errResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (e errResponse) Error() string {
	return e.Message
}

var errResponseBookEntryBlankFileds = errResponse{100, "all the fields - name, price and inventory - must be filled correctly."}
var errResponseBookNotFound = errResponse{101, "book not found"}
var errResponseBookEntryInvalidJSON = errResponse{102, "invalid json request."}
var errResponseIdInvalidFormat = errResponse{103, "the endpoint is not a valid format ID. Must be /books/{uuid}"}
var errResponseQueryPriceInvalidFormat = errResponse{104, "query parameter 'price' must be a float between 0 and 9999.99"}
var errResponseQuerySortByInvalid = errResponse{105, "query parameter 'sort_by' must be: name, price, inventory, created_at or updated_at. 'sort_direction' must be asc or desc."}
var errResponseQueryPageInvalid = errResponse{106, "query parameter 'page' must be an int starting in 1. 'page_size' must be an int beetween 1 and 30."}
var errResponseQueryPageOutOfRange = errResponse{107, "page out of range."}
