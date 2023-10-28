package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strings"
	"time"
)

func BuyTicket(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString := authParts[1]

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

		var ticketPurchase struct {
			HotelID        uint   `json:"hotel_id"`
			Night          int    `json:"night"`
			KodeVoucher    string `json:"kode_voucher"`
			UseAllPoints   bool   `json:"use_all_points"`
			UsedPoints     int    `json:"used_points"`
			Quantity       int    `json:"quantity"`
			CheckinBooking string `json:"checkin_booking"`
		}

		if err := c.Bind(&ticketPurchase); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var hotel model.Hotel
		hotelResult := db.First(&hotel, ticketPurchase.HotelID)
		if hotelResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Hotel not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		checkinBookingTime, err := time.Parse("2006-01-02", ticketPurchase.CheckinBooking)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid checkin_booking date format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		checkoutBookingTime := checkinBookingTime.AddDate(0, 0, ticketPurchase.Night)

		var discountPercentage int
		totalCost := hotel.Price * ticketPurchase.Night * ticketPurchase.Quantity

		pointsEarned := totalCost / 10000

		if ticketPurchase.KodeVoucher != "" {
			var promo model.Promo
			promoResult := db.Where("kode_voucher = ?", ticketPurchase.KodeVoucher).First(&promo)
			if promoResult.Error == nil {
				discountPercentage = promo.JumlahPotonganPersen
				if discountPercentage > 0 {
					discount := (totalCost * discountPercentage) / 100
					totalCost -= discount
				}
				pointsEarned = 0
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid kode voucher"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		usedPoints := 0

		if ticketPurchase.UseAllPoints {
			usedPoints = user.Points
			additionalDiscount := usedPoints * 1000
			if additionalDiscount <= totalCost {
				totalCost -= additionalDiscount
				user.Points = 0

				if err := db.Save(&user).Error; err != nil {
					errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user points"}
					return c.JSON(http.StatusInternalServerError, errorResponse)
				}
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough points to use"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		} else {
			usedPoints = ticketPurchase.UsedPoints
			additionalDiscount := usedPoints * 1000
			if additionalDiscount <= totalCost {
				totalCost -= additionalDiscount
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough points to use"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if hotel.AvailableRooms < ticketPurchase.Quantity {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough available rooms"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		hotel.AvailableRooms -= ticketPurchase.Quantity
		if err := db.Save(&hotel).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update available rooms"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Buat tiket dan simpan di database
		ticket := model.Ticket{
			HotelID:         hotel.ID,
			UserID:          user.ID,
			Night:           ticketPurchase.Night,
			UsedPoints:      usedPoints,
			TotalCost:       totalCost,
			InvoiceNumber:   helper.GenerateInvoiceNumber(),
			KodeVoucher:     ticketPurchase.KodeVoucher,
			Quantity:        ticketPurchase.Quantity,
			CheckinBooking:  &checkinBookingTime,
			CheckoutBooking: &checkoutBookingTime,
			PaidStatus:      false,        // Set paid_status menjadi false saat pembelian
			PointsEarned:    pointsEarned, // Simpan pointsEarned dalam tiket
		}

		if err := db.Create(&ticket).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create ticket"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Kirim email ke pengguna
		emailSubject := helper.GetEmailSubject(ticket)
		hotelName := hotel.Title
		hotelRoom := hotel.RoomType
		emailBody := helper.GetEmailBody(ticket, totalCost, hotelName, hotelRoom, ticketPurchase.KodeVoucher, pointsEarned, usedPoints)

		if err := helper.SendEmailToUser(user.Email, emailSubject, emailBody); err != nil {
			fmt.Println("Failed to send email:", err)
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send email to user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		pointMessage := "Points earned"
		if pointsEarned == 0 && ticketPurchase.KodeVoucher != "" {
			pointMessage = "Points not earned due to voucher"
		}

		// Setelah transaksi selesai, periksa paid_status pada tiket
		var updatedTicket model.Ticket
		result = db.Where("invoice_number = ?", ticket.InvoiceNumber).First(&updatedTicket)
		if result.Error == nil && updatedTicket.PaidStatus && !user.IsAdmin {
			// Jika paid_status berubah menjadi true dan pengguna bukan admin, tambahkan poin ke pengguna
			user.Points += pointsEarned

			if err := db.Save(&user).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user points"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		responseData := map[string]interface{}{
			"error":          false,
			"message":        "Ticket Hotel purchased successfully",
			"ticket_id":      ticket.ID,
			"invoice_number": ticket.InvoiceNumber,
			"total_cost":     totalCost,
			"kode_voucher":   ticketPurchase.KodeVoucher,
			"points_earned":  pointsEarned,
			"used_points":    usedPoints,
			"point_message":  pointMessage,
		}
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Ticket Hotel purchased successfully",
			"data":    responseData,
		}

		return c.JSON(http.StatusOK, response)
	}
}
