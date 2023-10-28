package helper

import (
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

func SendWelcomeEmail(userEmail, name, verificationToken string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	sender := smtpUsername
	recipient := userEmail
	subject := "Welcome to Hotelku Booking App"
	verificationLink := "https://hotelku-miniproject-403405.uc.r.appspot.com/verify?token=" + verificationToken
	emailBody := `
    <html>
    <head>
        <link href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
        <style>
            body {
                font-family: Arial, sans-serif;
                background-color: #f5f5f5;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #fff;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
                border-radius: 5px;
            }
            h1 {
                text-align: center;
                color: #333;
            }
            .message {
                background-color: #f9f9f9;
                padding: 15px;
                border: 1px solid #ddd;
            }
            p {
                font-size: 16px;
                margin-top: 10px;
            }
            strong {
                font-weight: bold;
            }
            .footer {
                text-align: center;
                margin-top: 20px;
                color: #666;
            }
            .btn-verify-email {
                background-color: #ff6600;
                color: #fff;
                padding: 10px 20px;
                border-radius: 5px;
                text-decoration: none;
                display: block;
                text-align: center;
                margin: 20px auto;
            }
            .btn-verify-email:hover {
                background-color: #ff3300;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Welcome to Our Hotel Booking App</h1>
            <div class="message">
                <p>Hello, <strong>` + name + `</strong>,</p>
                <p>Thank you for signing up with our hotel booking app. You're now part of our community!</p>
                <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
                <p><strong>Support Team:</strong> <a href="mailto:altaminiproject@gmail.com">altaminiproject@gmail.com</a></p>
                <a href="` + verificationLink + `" class="btn btn-verify-email">Verify Email</a>
            </div>
            <div class="footer">
                <p>&copy; 2023 Hotelku. All rights reserved.</p>
            </div>
        </div>
    </body>
    </html>
    `

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", emailBody)

	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
