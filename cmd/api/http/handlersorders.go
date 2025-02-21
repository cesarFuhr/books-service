package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/books-service/cmd/api/book"
	"github.com/google/uuid"
)

/* Addresses a call to "/order" according to the requested action.  */
func (h *BookHandler) order(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.requestTimeout))
	defer cancel()
	r = r.WithContext(ctx)

	method := r.Method
	switch method {
	case http.MethodPost:
		h.createOrder(w, r)
		return
	case http.MethodGet:
		h.listOrderItems(w, r)
		return
	case http.MethodPut:
		h.updateOrder(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

type NewOrderEntry struct {
	UserID uuid.UUID `json:"user_id"`
}

/* Validates the entry, then creates an empty order. */
func (h *BookHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var newOrderEntry NewOrderEntry
	err := json.NewDecoder(r.Body).Decode(&newOrderEntry)
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseEntryInvalidJSON.Code,
			Message: book.ErrResponseEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	if newOrderEntry.UserID.String() == "" {
		responseJSON(w, http.StatusBadRequest, book.ErrResponseNewOrderEntryBlankFields)
	}

	newOrder, err := h.bookService.CreateOrder(r.Context(), newOrderEntry.UserID)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, orderToResponse(newOrder))
}

type UpdateOrderEntry struct {
	OrderID        uuid.UUID `json:"order_id"`
	BookID         uuid.UUID `json:"book_id"`
	BookUnitsToAdd int       `json:"book_units_to_add"`
}

/* Validates the entry, then updates the order adding or removing books. */
func (h *BookHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
	var updateOrderEntry UpdateOrderEntry
	err := json.NewDecoder(r.Body).Decode(&updateOrderEntry)
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseEntryInvalidJSON.Code,
			Message: book.ErrResponseEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	err = FilledUpdtOrderFields(updateOrderEntry) //Verify if all entry fields are filled.
	if err != nil {
		responseJSON(w, http.StatusBadRequest, err)
		return
	}

	reqOrder := updateOrderToUpdateReq(updateOrderEntry)

	updatedOrder, err := h.bookService.UpdateOrderTx(r.Context(), reqOrder) //Update the stored order
	if err != nil {
		handleError(fmt.Errorf("error on call to UpdateOrderTx: %w ", err), w, r)
		return
	}

	responseJSON(w, http.StatusOK, orderToResponse(updatedOrder))
}

/* Verifies if all UpdateOrder entry fields are filled and returns a warning message if so. */
func FilledUpdtOrderFields(updtOrderEntry UpdateOrderEntry) error {
	if updtOrderEntry.OrderID.String() == "" {
		return book.ErrResponseUpdateOrderEntryBlankFields
	}
	if updtOrderEntry.BookID.String() == "" {
		return book.ErrResponseUpdateOrderEntryBlankFields
	}
	if updtOrderEntry.BookUnitsToAdd == 0 { //If this value comes 0, than nothing changes, so it's not valid.
		return book.ErrResponseUpdateOrderEntryBlankFields
	}

	return nil
}

/* Converts from UpdateOrderEntry type to UpdateOrderRequest type, with no json tags. */
func updateOrderToUpdateReq(o UpdateOrderEntry) book.UpdateOrderRequest {
	return book.UpdateOrderRequest{
		OrderID:        o.OrderID,
		BookID:         o.BookID,
		BookUnitsToAdd: o.BookUnitsToAdd,
	}
}

type OrderResponse struct {
	OrderID     uuid.UUID           `json:"order_id"`
	PurchaserID uuid.UUID           `json:"purchaser_id"`
	OrderStatus string              `json:"order_status"`
	TotalPrice  float32             `json:"total_price"`
	Items       []OrderItemResponse `json:"order_items"`
}

/*Copy the fields of an order object to an http layer struct with json tags*/
func orderToResponse(o book.Order) OrderResponse {
	items := []OrderItemResponse{}
	for _, item := range o.Items {
		items = append(items, orderItemToResponse(item))
	}

	return OrderResponse{
		OrderID:     o.OrderID,
		PurchaserID: o.PurchaserID,
		OrderStatus: o.OrderStatus,
		TotalPrice:  o.TotalPrice,
		Items:       items,
	}
}

type OrderItemResponse struct {
	BookID           uuid.UUID `json:"book_id"`
	BookName         string    `json:"book_name"`
	BookUnits        int       `json:"book_units"`
	BookPriceAtOrder *float32  `json:"book_price"`
}

/*Copy the fields of an orderItem object to an http layer struct with json tags*/
func orderItemToResponse(i book.OrderItem) OrderItemResponse {
	return OrderItemResponse{
		BookID:           i.BookID,
		BookName:         i.BookName,
		BookUnits:        i.BookUnits,
		BookPriceAtOrder: i.BookPriceAtOrder,
	}
}

type ListItemsEntry struct {
	OrderID uuid.UUID `json:"order_id"`
}

/* Validates the entry, then creates an empty order. */
func (h *BookHandler) listOrderItems(w http.ResponseWriter, r *http.Request) {
	var listItemsEntry ListItemsEntry
	err := json.NewDecoder(r.Body).Decode(&listItemsEntry)
	if err != nil {
		log.Println(err)
		errR := book.ErrResponse{
			Code:    book.ErrResponseEntryInvalidJSON.Code,
			Message: book.ErrResponseEntryInvalidJSON.Message + err.Error(),
		}
		responseJSON(w, http.StatusBadRequest, errR)
		return
	}

	if listItemsEntry.OrderID.String() == "" {
		responseJSON(w, http.StatusBadRequest, book.ErrResponseListOrderItemsEntryBlankFields)
	}

	order, err := h.bookService.ListOrderItems(r.Context(), listItemsEntry.OrderID)
	if err != nil {
		handleError(err, w, r)
		return
	}

	responseJSON(w, http.StatusOK, orderToResponse(order))
}
