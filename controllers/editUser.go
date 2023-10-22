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

func EditUser(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		userID := c.Param("id")

		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if user.Username != username {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized to edit this user"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var updateUser model.User
		if err := c.Bind(&updateUser); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updateUser.Name != "" {
			user.Name = updateUser.Name
		}

		if updateUser.Username != "" {
			user.Username = updateUser.Username
		}
		if updateUser.Email != "" {
			user.Email = updateUser.Email
		}
		if updateUser.PhoneNumber != "" {
			user.PhoneNumber = updateUser.PhoneNumber
		}
		if updateUser.IsAdmin != user.IsAdmin {
			user.IsAdmin = updateUser.IsAdmin
		}

		if err := db.Save(&user).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		userResponse := helper.EditUserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Username:    user.Username,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "User updated successfully",
			"user":    userResponse,
		})
	}
}
