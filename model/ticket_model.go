package model

import "time"

type Ticket struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	HotelID         uint       `json:"hotel_id"`
	UserID          uint       `json:"user_id"`
	Night           int        `json:"night"`
	KodeVoucher     string     `json:"kode_voucher"`
	UsedPoints      int        `json:"used_points"`
	UseAllPoints    bool       `json:"use_all_points"`
	TotalCost       int        `json:"total_cost"`
	InvoiceNumber   string     `json:"invoice_number"`
	Quantity        int        `json:"quantity"`
	CheckinBooking  *time.Time `json:"checkin_booking"`  // Tanggal check-in
	CheckoutBooking *time.Time `json:"checkout_booking"` // Tanggal check-out
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       time.Time
}
