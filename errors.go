package main

var errResonseCreateBookBadInput = errorResponse{100, "All the fields - name, price and inventory - must be filled correctly."}
var errResonseCreateBookNameConflict = errorResponse{101, "There is already a book with this name on database:"}
