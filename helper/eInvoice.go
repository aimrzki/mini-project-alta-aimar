package helper

import (
	"fmt"
	"myproject/model"
)

// GetEmailSubject mengembalikan subjek email
func GetEmailSubject(ticket model.Ticket) string {
	return "Pembelian Tiket Hotel Berhasil - Invoice No: " + ticket.InvoiceNumber
}

// GetEmailBody mengembalikan body email
func GetEmailBody(ticket model.Ticket, totalCost int, hotelName, hotelRoom string, kodeVoucher string, pointsEarned, usedPoints int) string {
	emailBody := "<html><head><style>"
	emailBody += "body {font-family: Arial, sans-serif;}"
	emailBody += ".container {max-width: 600px; margin: 0 auto; padding: 20px;}"
	emailBody += "h1 {font-size: 24px; margin: 0;}"
	emailBody += "p {font-size: 16px; margin-top: 10px;}"
	emailBody += "strong {font-weight: bold;}"
	emailBody += ".invoice-details {background-color: #f5f5f5; padding: 10px; margin-top: 20px;}"
	emailBody += "hr {border: 1px solid #ccc; margin: 20px 0;}"
	emailBody += "</style></head><body>"
	emailBody += "<div class='container'>"
	emailBody += "<h1>Pembelian Tiket Hotel Berhasil</h1>"
	emailBody += "<p>Terima kasih telah melakukan pembelian tiket hotel. Berikut adalah rincian pembelian Anda:</p>"
	emailBody += "<div class='invoice-details'>"
	emailBody += "<p><strong>Nomor Invoice:</strong> " + ticket.InvoiceNumber + "</p>"
	emailBody += "<p><strong>Nama Hotel:</strong> " + hotelName + "</p>"
	emailBody += "<p><strong>Tipe Kamar:</strong> " + hotelRoom + "</p>"
	emailBody += "<p><strong>Quantity:</strong> " + fmt.Sprintf("%d", ticket.Quantity) + "</p>" // Menambahkan quantity di sini
	emailBody += "<p><strong>Night:</strong> " + fmt.Sprintf("%d", ticket.Night) + "</p>"
	emailBody += "<p><strong>Check-in Booking:</strong> " + ticket.CheckinBooking.Format("2006-01-02") + "</p>"
	emailBody += "<p><strong>Checkout Booking:</strong> " + ticket.CheckoutBooking.Format("2006-01-02") + "</p>"
	emailBody += "<p><strong>Total Harga:</strong> Rp. " + fmt.Sprintf("%d", totalCost) + "</p>"

	if kodeVoucher != "" {
		emailBody += "<p><strong>Kode Voucher:</strong> " + kodeVoucher + "</p>"
	}

	emailBody += "<p><strong>Points Earned:</strong> " + fmt.Sprintf("%d", pointsEarned) + "</p>"
	emailBody += "<p><strong>Used Points:</strong> " + fmt.Sprintf("%d", usedPoints) + "</p>"
	emailBody += "</div>"
	emailBody += "<hr>"
	emailBody += "<p>Terima kasih atas pembelian Anda!</p>"
	emailBody += "<p>Kami berharap Anda memiliki pengalaman yang menyenangkan di hotel kami.</p>"
	emailBody += "<div style='text-align: left; font-size: 14px; margin-top: 20px;'>"
	emailBody += "<p style='font-weight: bold; margin: 0;'>Hormat Saya</p>"
	emailBody += "<p style='font-size: 12px; margin: 0;'>Muhammad Aimar Rizki Utama 💙</p>"
	emailBody += "<p style='font-size: 12px; margin: 0;'>CEO Hotelku Indonesia</p>"
	emailBody += "</div>"
	emailBody += "</body></html>"

	return emailBody
}
