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
	e.GET("/verify", controllers.VerifyEmail(db))
	e.POST("/signin", controllers.Signin(db, secretKey))
	e.POST("/admin/signin", controllers.AdminSignin(db, secretKey))
	e.POST("/hotel/create", controllers.CreateHotel(db, secretKey))
	e.POST("/createpromo", controllers.CreatePromo(db, secretKey))
	e.PUT("/admin/hotel/:id", controllers.EditHotel(db, secretKey))
	e.GET("/admin/user", controllers.GetAllUsersByAdmin(db, secretKey))
	e.GET("/admin/ticket", controllers.GetAllTicketsByAdmin(db, secretKey))
	e.GET("/admin/ticket/invoice", controllers.GetTicketByInvoiceNumber(db, secretKey))
	e.DELETE("/admin/hotel/:id", controllers.DeleteHotel(db, secretKey))
	e.DELETE("/admin/user/:id", controllers.DeleteUserByAdmin(db, secretKey))
	e.GET("/user/:user_id", controllers.GetUserDataByID(db, secretKey))
	e.GET("/hotel", controllers.GetHotels(db, secretKey))
	e.GET("/hotel/:id", controllers.GetHotelByID(db, secretKey))
	e.GET("/user/promo", controllers.GetPromos(db, secretKey))
}

func getSecretKeyFromEnv() string {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY tidak ditemukan di .env")
	}
	return secretKey
}
