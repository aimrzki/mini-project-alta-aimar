package model

type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Username    string `gorm:"uniqueIndex" json:"username"`
	Email       string `gorm:"uniqueIndex" json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `gorm:"uniqueIndex" json:"phone_number"`
	Points      int    `json:"points"`                       // Menambahkan field untuk menyimpan poin pengguna
	IsAdmin     bool   `gorm:"default:false" json:"isAdmin"` // Menambahkan nilai default
}

// Buat struct untuk permintaan perubahan kata sandi
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}
