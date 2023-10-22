package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"strings"
)

func CreateHotel(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser model.User
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Anda bukan admin!"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Hanya Admin yang dapat menambahkan hotel"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		title := c.FormValue("title")
		location := c.FormValue("location")
		description := c.FormValue("description")
		price := c.FormValue("price")
		availableRooms := c.FormValue("available_rooms")
		roomType := c.FormValue("room_type")
		guestCount := c.FormValue("guest_count")

		priceInt, err := strconv.Atoi(price)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid price value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		availableRoomsInt, err := strconv.Atoi(availableRooms)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid available_rooms value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		guestCountInt, err := strconv.Atoi(guestCount)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid guest_count value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		imageFile, err := c.FormFile("image")
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Failed to read image data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if !helper.IsImageFile(imageFile) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid file type. Only image files are allowed."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if helper.IsFileSizeExceeds(imageFile, 5*1024*1024) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "File size exceeds the allowed limit (5MB)."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		src, err := imageFile.Open()
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Failed to open image file"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		defer src.Close()

		imageData, err := io.ReadAll(src)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Failed to read image data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		createdHotel := model.Hotel{
			Title:          title,
			Location:       location,
			Description:    description,
			Price:          priceInt,
			AvailableRooms: availableRoomsInt,
			RoomType:       roomType,
			GuestCount:     guestCountInt,
			UserID:         adminUser.ID,
		}

		if err := db.Create(&createdHotel).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create hotel"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		imageName := fmt.Sprintf("hotel_%d.jpg", createdHotel.ID)
		imageURL, err := helper.UploadImageToGCS(imageData, imageName)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to upload image to GCS"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		createdHotel.PhotoHotel = imageURL

		if err := db.Save(&createdHotel).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create hotel"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "Hotel created successfully",
			"hotelData": createdHotel,
		})
	}
}
