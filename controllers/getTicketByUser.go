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

func GetTicketsByUser(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
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

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan ID pengguna dari token
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengambil tiket yang telah dibeli oleh pengguna berdasarkan UserID
		var tickets []model.Ticket
		result = db.Where("user_id = ?", user.ID).Find(&tickets)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user's tickets"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Membuat respons dengan data tiket yang telah dibeli
		var ticketDetails []map[string]interface{}
		for _, ticket := range tickets {
			// Mengambil detail event berdasarkan EventID yang ada pada tiket
			var hotel model.Hotel
			eventResult := db.First(&hotel, ticket.HotelID)
			if eventResult.Error != nil {
				// Handle jika event tidak ditemukan
				continue
			}

			// Menambahkan informasi kode voucher yang digunakan
			var kodeVoucher string
			if ticket.KodeVoucher != "" {
				kodeVoucher = ticket.KodeVoucher
			}

			// Menambahkan nomor invoice ke detail tiket
			ticketDetail := map[string]interface{}{
				"user_id":        ticket.UserID,
				"hotel_id":       ticket.HotelID,
				"hotel_name":     hotel.Title,
				"room_type":      hotel.RoomType,
				"guest_count":    hotel.GuestCount,
				"night":          ticket.Night,
				"total_cost":     ticket.TotalCost,
				"invoice_number": ticket.InvoiceNumber,
				"kode_voucher":   kodeVoucher,
				"quantity":       ticket.Quantity,   // Menambahkan quantity ke detail tiket
				"paid":           ticket.PaidStatus, // Menambahkan status pembayaran ke detail tiket
			}

			// Menambahkan objek tiket ke daftar ticketDetails
			ticketDetails = append(ticketDetails, ticketDetail)
		}

		// Mengembalikan respons dengan detail tiket yang telah dibeli
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "User's tickets retrieved successfully",
			"hotel_data": ticketDetails,
		})
	}
}
