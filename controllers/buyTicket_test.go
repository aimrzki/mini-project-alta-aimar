package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func SigninWithMock(db *gorm.DB, secretKey []byte, mockSendLoginNotification *MockSendLoginNotification) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user model.User
		if err := c.Bind(&user); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username ada dalam database
		var existingUser model.User
		result := db.Where("username = ?", user.Username).First(&existingUser)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username or password"}
				return c.JSON(http.StatusUnauthorized, errorResponse)
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Membandingkan password yang dimasukkan dengan password yang di-hash
		err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username or password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if existingUser.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied. Admin cannot use this endpoint."}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		if !existingUser.IsVerified {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Account not verified. Please verify your email before logging in."}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Gantilah implementasi SendLoginNotification dengan mock
		if err := mockSendLoginNotification.SendLoginNotification(existingUser.Email, existingUser.Name); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send notification email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(existingUser.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "User login successful", "token": tokenString, "id": existingUser.ID})
	}
}

func BuyTicketWithMock(db *gorm.DB, secretKey []byte, mockSendLoginNotification *MockSendLoginNotification) echo.HandlerFunc {
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

func TestBuyTicketWithMock(t *testing.T) {
	// Inisialisasi Echo framework dan database (gunakan database in-memory untuk pengujian)
	dsn := "root:@tcp(localhost:3306)/testing?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	e := echo.New()

	// Simulasikan login pengguna dan mendapatkan token
	loginPayload := map[string]string{
		"username": "aimrzki",
		"password": "user123",
	}
	loginJSON, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(loginJSON))
	reqLogin.Header.Set("Content-Type", "application/json")
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)
	SigninWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cLogin)

	// Periksa status kode dari respons login
	if recLogin.Code != http.StatusOK {
		t.Fatalf("Login failed with status code: %d", recLogin.Code)
	}

	// Ambil token dari respons login
	var loginResponse map[string]interface{}
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := loginResponse["token"].(string)
	if !ok {
		t.Fatalf("Token not found in login response")
	}

	// Gunakan token untuk membeli tiket
	buyTicketPayload := map[string]interface{}{
		"hotel_id":        1,
		"night":           2,
		"kode_voucher":    "tahunbarualta50",
		"use_all_points":  false,
		"quantity":        1,
		"checkin_booking": "2023-11-01",
	}
	buyTicketJSON, _ := json.Marshal(buyTicketPayload)
	reqBuyTicket := httptest.NewRequest(http.MethodPost, "/user/buy", bytes.NewBuffer(buyTicketJSON))
	reqBuyTicket.Header.Set("Content-Type", "application/json")
	reqBuyTicket.Header.Set("Authorization", "Bearer "+token)
	recBuyTicket := httptest.NewRecorder()
	cBuyTicket := e.NewContext(reqBuyTicket, recBuyTicket)

	BuyTicketWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cBuyTicket)
	fmt.Printf("Buy Ticket Response: %s\n", recBuyTicket.Body.String())

	if recBuyTicket.Code != http.StatusOK {
		t.Fatalf("Failed to purchase ticket with status code: %d", recBuyTicket.Code)
	}
}

func TestFailedBuyTicketWithMock(t *testing.T) {
	// Inisialisasi Echo framework dan database (gunakan database in-memory untuk pengujian)
	dsn := "root:@tcp(localhost:3306)/testing?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	e := echo.New()

	// Simulasikan login pengguna dan mendapatkan token
	loginPayload := map[string]string{
		"username": "aimrzki",
		"password": "user123",
	}
	loginJSON, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(loginJSON))
	reqLogin.Header.Set("Content-Type", "application/json")
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)
	SigninWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cLogin)

	// Periksa status kode dari respons login
	if recLogin.Code != http.StatusOK {
		t.Fatalf("Login failed with status code: %d", recLogin.Code)
	}

	// Ambil token dari respons login
	var loginResponse map[string]interface{}
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := loginResponse["token"].(string)
	if !ok {
		t.Fatalf("Token not found in login response")
	}

	// Gunakan token untuk mencoba pembelian tiket yang gagal
	invalidBuyTicketPayload := map[string]interface{}{
		"hotel_id":        1,
		"night":           -2, // Jumlah malam negatif
		"kode_voucher":    "voucher_tidak_valid",
		"use_all_points":  false,
		"quantity":        10, // Jumlah kamar yang melebihi stok hotel
		"checkin_booking": "2023-10-01",
	}
	invalidBuyTicketJSON, _ := json.Marshal(invalidBuyTicketPayload)
	reqInvalidBuyTicket := httptest.NewRequest(http.MethodPost, "/user/buy", bytes.NewBuffer(invalidBuyTicketJSON))
	reqInvalidBuyTicket.Header.Set("Content-Type", "application/json")
	reqInvalidBuyTicket.Header.Set("Authorization", "Bearer "+token)
	recInvalidBuyTicket := httptest.NewRecorder()
	cInvalidBuyTicket := e.NewContext(reqInvalidBuyTicket, recInvalidBuyTicket)

	BuyTicketWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cInvalidBuyTicket)
	fmt.Printf("Invalid Buy Ticket Response: %s\n", recInvalidBuyTicket.Body.String())

	if recInvalidBuyTicket.Code == http.StatusOK {
		t.Fatalf("Purchased invalid ticket successfully with status code: %d", recInvalidBuyTicket.Code)
	}
}

func TestTokenValidationWithMock(t *testing.T) {
	// Inisialisasi Echo framework dan database (gunakan database in-memory untuk pengujian)
	dsn := "root:@tcp(localhost:3306)/testing?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	e := echo.New()

	// Coba membeli tiket tanpa token otentikasi
	buyTicketPayload := map[string]interface{}{
		"hotel_id":        1,
		"night":           2,
		"kode_voucher":    "tahunbarualta50",
		"use_all_points":  false,
		"quantity":        1,
		"checkin_booking": "2023-11-01",
	}
	buyTicketJSON, _ := json.Marshal(buyTicketPayload)
	reqBuyTicket := httptest.NewRequest(http.MethodPost, "/user/buy", bytes.NewBuffer(buyTicketJSON))
	reqBuyTicket.Header.Set("Content-Type", "application/json")
	recBuyTicket := httptest.NewRecorder()
	cBuyTicket := e.NewContext(reqBuyTicket, recBuyTicket)

	BuyTicketWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cBuyTicket)
	fmt.Printf("Buy Ticket Without Token Response: %s\n", recBuyTicket.Body.String())

	if recBuyTicket.Code != http.StatusUnauthorized {
		t.Fatalf("Purchased ticket without token failed with status code: %d", recBuyTicket.Code)
	}

	// Coba membeli tiket dengan token otentikasi tidak valid
	invalidTokenBuyTicketPayload := map[string]interface{}{
		"hotel_id":        1,
		"night":           2,
		"kode_voucher":    "tahunbarualta50",
		"use_all_points":  false,
		"quantity":        1,
		"checkin_booking": "2023-11-01",
	}
	invalidTokenBuyTicketJSON, _ := json.Marshal(invalidTokenBuyTicketPayload)
	reqInvalidTokenBuyTicket := httptest.NewRequest(http.MethodPost, "/user/buy", bytes.NewBuffer(invalidTokenBuyTicketJSON))
	reqInvalidTokenBuyTicket.Header.Set("Content-Type", "application/json")
	reqInvalidTokenBuyTicket.Header.Set("Authorization", "Bearer InvalidToken")
	recInvalidTokenBuyTicket := httptest.NewRecorder()
	cInvalidTokenBuyTicket := e.NewContext(reqInvalidTokenBuyTicket, recInvalidTokenBuyTicket)

	BuyTicketWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cInvalidTokenBuyTicket)
	fmt.Printf("Buy Ticket With Invalid Token Response: %s\n", recInvalidTokenBuyTicket.Body.String())

	if recInvalidTokenBuyTicket.Code != http.StatusUnauthorized {
		t.Fatalf("Purchased ticket with invalid token failed with status code: %d", recInvalidTokenBuyTicket.Code)
	}
}

func TestInvalidCheckinTimeWithMock(t *testing.T) {
	// Inisialisasi Echo framework dan database (gunakan database in-memory untuk pengujian)
	dsn := "root:@tcp(localhost:3306)/testing?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	e := echo.New()

	// Simulasikan login pengguna dan mendapatkan token
	loginPayload := map[string]string{
		"username": "aimrzki",
		"password": "user123",
	}
	loginJSON, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(loginJSON))
	reqLogin.Header.Set("Content-Type", "application/json")
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)
	SigninWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cLogin)

	// Periksa status kode dari respons login
	if recLogin.Code != http.StatusOK {
		t.Fatalf("Login failed with status code: %d", recLogin.Code)
	}

	// Ambil token dari respons login
	var loginResponse map[string]interface{}
	if err := json.Unmarshal(recLogin.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := loginResponse["token"].(string)
	if !ok {
		t.Fatalf("Token not found in login response")
	}

	// Coba membeli tiket dengan format waktu check-in yang tidak valid
	invalidCheckinTimePayload := map[string]interface{}{
		"hotel_id":        1,
		"night":           2,
		"kode_voucher":    "tahunbarualta50",
		"use_all_points":  false,
		"quantity":        1,
		"checkin_booking": "InvalidTime", // Format waktu tidak valid
	}
	invalidCheckinTimeJSON, _ := json.Marshal(invalidCheckinTimePayload)
	reqInvalidCheckinTime := httptest.NewRequest(http.MethodPost, "/user/buy", bytes.NewBuffer(invalidCheckinTimeJSON))
	reqInvalidCheckinTime.Header.Set("Content-Type", "application/json")
	reqInvalidCheckinTime.Header.Set("Authorization", "Bearer "+token)
	recInvalidCheckinTime := httptest.NewRecorder()
	cInvalidCheckinTime := e.NewContext(reqInvalidCheckinTime, recInvalidCheckinTime)

	BuyTicketWithMock(db, []byte("secret-key"), &MockSendLoginNotification{})(cInvalidCheckinTime)
	fmt.Printf("Buy Ticket With Invalid Check-in Time Response: %s\n", recInvalidCheckinTime.Body.String())

	if recInvalidCheckinTime.Code == http.StatusBadRequest {
		// Pastikan respons mengandung pesan kesalahan yang sesuai
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(recInvalidCheckinTime.Body.Bytes(), &errorResponse); err == nil {
			message, ok := errorResponse["message"].(string)
			if ok && strings.Contains(message, "Invalid checkin_booking date format") {
				return // Tes berhasil jika pesan kesalahan sesuai dengan format yang diharapkan
			}
		}
	}

	t.Fatalf("Test with invalid check-in time failed with status code: %d", recInvalidCheckinTime.Code)
}

type MockSendLoginNotification struct{}

func (m *MockSendLoginNotification) SendLoginNotification(email, name string) error {
	return nil
}
