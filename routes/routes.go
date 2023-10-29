package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"myproject/controllers"
	"net/http"
	"os"
)

func ServeHTML(c echo.Context) error {
	htmlData, err := ioutil.ReadFile("index.html")
	if err != nil {
		return err
	}
	return c.HTML(http.StatusOK, string(htmlData))
}

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	e.Use(Logger())
	secretKey := []byte(getSecretKeyFromEnv())
	// Menggunakan routes yang telah dipisahkan
	e.GET("/", ServeHTML)
	e.POST("/signup", controllers.Signup(db, secretKey))
	e.GET("/verify", controllers.VerifyEmail(db))
	e.POST("/signin", controllers.Signin(db, secretKey))
	e.POST("/admin/signin", controllers.AdminSignin(db, secretKey))
	e.PUT("/user/change-password/:id", controllers.ChangePassword(db, secretKey))
	e.POST("/hotel/create", controllers.CreateHotel(db, secretKey))
	e.GET("/hotel", controllers.GetHotels(db, secretKey))
	e.GET("/hotel/:id", controllers.GetHotelByID(db, secretKey))
	e.PUT("/user/:id", controllers.EditUser(db, secretKey))
	e.DELETE("/admin/user/:id", controllers.DeleteUserByAdmin(db, secretKey))
	e.GET("/admin/user", controllers.GetAllUsersByAdmin(db, secretKey))
	e.POST("/user/buy", controllers.BuyTicket(db, secretKey))
	e.GET("/user/hotel", controllers.GetTicketsByUser(db, secretKey))
	e.POST("/createpromo", controllers.CreatePromo(db, secretKey))
	e.GET("/user/promo", controllers.GetPromos(db, secretKey))
	e.GET("/admin/ticket/:invoiceNumber", controllers.GetTicketByInvoiceNumber(db, secretKey))
	e.PUT("/admin/ticket/:invoiceId", controllers.UpdatePaidStatus(db, secretKey))
	e.GET("/user/points", controllers.GetUserPoints(db, secretKey))
	e.GET("/user/:user_id", controllers.GetUserDataByID(db, secretKey))
	e.PUT("/admin/hotel/:id", controllers.EditHotel(db, secretKey))
	e.GET("/admin/ticket", controllers.GetAllTicketsByAdmin(db, secretKey))
	e.DELETE("/admin/hotel/:id", controllers.DeleteHotel(db, secretKey))
	hotelUsecase := controllers.NewHotelUsecase() // Inisialisasi use case
	e.POST("/chatbot/recommend-hotel", func(c echo.Context) error {
		return controllers.RecommendHotel(c, hotelUsecase) // Panggil fungsi dengan use case
	})
}

func getSecretKeyFromEnv() string {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY tidak ditemukan di .env")
	}
	return secretKey
}
