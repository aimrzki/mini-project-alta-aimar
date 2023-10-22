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

func DeleteHotel(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
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

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan informasi admin yang diautentikasi
		var adminUser model.User
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Anda bukan admin!"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah admin yang diautentikasi memiliki status IsAdmin yang true
		if !adminUser.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Hanya Admin yang dapat menghapus hotel"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan ID hotel dari parameter URL
		hotelID := c.Param("id")

		// Cek apakah hotel dengan ID tersebut ada dalam basis data
		var existingHotel model.Hotel
		if err := db.Where("id = ?", hotelID).First(&existingHotel).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Hotel not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Menghapus hotel dari basis data
		if err := db.Delete(&existingHotel).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete hotel"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengembalikan respons sukses jika berhasil
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Hotel deleted successfully",
		})
	}
}
