package model

import "time"

type Hotel struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Title          string     `json:"title"`
	Location       string     `json:"location"`
	Description    string     `json:"description"`
	Price          int        `json:"price"`
	UserID         uint       `json:"user_id"` // ID pengguna yang membuat event
	AvailableRooms int        `json:"available_rooms"`
	RoomType       string     `json:"room_type"`   // Tipe Kamar (contoh: "Standard", "Deluxe", dll.)
	GuestCount     int        `json:"guest_count"` // Jumlah Tamu
	PhotoHotel     string     `json:"photo_hotel"` // Tambahkan field untuk link URL gambar hotel
	CreatedAt      *time.Time `json:"created_at"`  // Kolom created_at yang diharapkan tipe data *time.Time
	UpdatedAt      time.Time
}

// Metode untuk mengurangi jumlah tiket yang tersedia
func (e *Hotel) DecrementTickets(quantity int) {
	if e.AvailableRooms >= quantity {
		e.AvailableRooms -= quantity
	}
}
