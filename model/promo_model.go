package model

type Promo struct {
	ID                   uint   `gorm:"primaryKey" json:"id"`
	Title                string `json:"title"`
	KodeVoucher          string `gorm:"uniqueIndex" json:"kode_voucher"`
	JumlahPotonganPersen int    `json:"jumlah_potongan_persen"`
}
