package model

type User struct {
	ID                uint    `gorm:"primaryKey" json:"id"`
	Name              string  `json:"name"`
	Username          string  `gorm:"uniqueIndex;size:255" json:"username"`
	Email             string  `gorm:"uniqueIndex;size:255" json:"email"`
	Password          string  `json:"password"`
	PhoneNumber       string  `gorm:"uniqueIndex;size:255" json:"phone_number"`
	Points            int     `json:"points"`
	IsAdmin           bool    `gorm:"default:false" json:"isAdmin"`
	IsVerified        bool    `gorm:"default:false" json:"is_verified"`
	VerificationToken string  `json:"verification_token"`
	Hotels            []Hotel `gorm:"foreignKey:UserID" json:"hotels"`
}

// Buat struct untuk permintaan perubahan kata sandi
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}
