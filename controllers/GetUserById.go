package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"strings"
)

func GetUserDataByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid user ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var authUser model.User
		if err := db.Where("username = ?", username).First(&authUser).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if authUser.IsAdmin || uint(userID) == authUser.ID {
			var user model.User
			if err := db.First(&user, uint(userID)).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}

			userResponse := helper.UserResponse{
				ID:          user.ID,
				Name:        user.Name,
				Username:    user.Username,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				Points:      user.Points,
				IsVerified:  user.IsVerified,
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    http.StatusOK,
				"error":   false,
				"message": "User data retrieved successfully",
				"user":    userResponse,
			})
		}

		errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
		return c.JSON(http.StatusForbidden, errorResponse)
	}
}
