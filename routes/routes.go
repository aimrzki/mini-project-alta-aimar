package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"log"
	"myproject/controllers"
	"os"
)

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	e.Use(Logger())
	secretKey := []byte(getSecretKeyFromEnv())
	// Menggunakan routes yang telah dipisahkan
	e.POST("/signup", controllers.Signup(db, secretKey))
	e.POST("/signin", controllers.Signin(db, secretKey))
	e.POST("/admin/signin", controllers.AdminSignin(db, secretKey))
}

func getSecretKeyFromEnv() string {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY tidak ditemukan di .env")
	}
	return secretKey
}
