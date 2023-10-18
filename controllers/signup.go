package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func Signup(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user model.User
		if err := c.Bind(&user); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username sudah ada dalam database
		var existingUser model.User
		result := db.Where("username = ?", user.Username).First(&existingUser)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Username already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengecek apakah email sudah ada dalam database
		result = db.Where("email = ?", user.Email).First(&existingUser)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Email already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengecek apakah phone number sudah ada dalam database
		result = db.Where("phone_number = ?", user.PhoneNumber).First(&existingUser)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Phone number already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check phone number"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Meng-hash password dengan bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menyimpan data pengguna ke database
		user.Password = string(hashedPassword)
		db.Create(&user)

		// Hapus password dari struct
		user.Password = ""

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(user.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if err := helper.SendWelcomeEmail(user.Email, user.Username); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menyertakan ID pengguna dalam respons
		return c.JSON(http.StatusOK, map[string]interface{}{"message": "User created successfully", "token": tokenString, "id": user.ID})
	}
}
