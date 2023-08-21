package main

type errResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

var errResponseCreateBookBlankFileds = errResponse{100, "all the fields - name, price and inventory - must be filled correctly."}
var errResponseCreateBookNameConflict = errResponse{101, "there is already a book with this name on database."}
var errResponseCreateBookInvalidJSON = errResponse{102, "invalid json request."}
var errResponseIdInvalidFormat = errResponse{103, "the endpoint is not a valid format ID. Must be /books/{uuid}"}
var errResponseQueryPriceInvalidFormat = errResponse{104, "query parameter price must be a float between 0 and 9999.99"}