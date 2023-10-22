package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strings"
)

func GetTicketByInvoiceNumber(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if !user.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access forbidden for non-admin users"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Dapatkan invoice_number dari body request
		var requestBody struct {
			InvoiceNumber string `json:"invoice_number"`
		}

		if err := c.Bind(&requestBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		invoiceNumber := requestBody.InvoiceNumber

		var ticket model.Ticket
		result = db.Where("invoice_number = ?", invoiceNumber).First(&ticket)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Ticket not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var hotel model.Hotel
		hotelResult := db.First(&hotel, ticket.HotelID)
		if hotelResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch event data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		ticketDetail := map[string]interface{}{
			"ticket_id":      ticket.ID,
			"user_id":        ticket.UserID,
			"hotel_id":       ticket.HotelID,
			"hotel_title":    hotel.Title,
			"room_type":      hotel.RoomType,
			"guest_count":    hotel.GuestCount,
			"night":          ticket.Night,
			"total_cost":     ticket.TotalCost,
			"invoice_number": ticket.InvoiceNumber,
			"kode_voucher":   ticket.KodeVoucher,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Ticket details retrieved successfully",
			"ticket_data": ticketDetail,
		})
	}
}
