package utils

import (
	"os"

	"github.com/mailgun/mailgun-go"
	"github.com/sirupsen/logrus"
)

//SendOTP ...
func SendOTP(to string, otp string) bool {
	//mailgun config
	mg := mailgun.NewMailgun(os.Getenv("DOMAIN"), os.Getenv("MAILGUN_API_KEY"))

	sender := ""
	subject := "Please Verify Your Email"
	body := "Your verification code is: " + otp + ". Valid for 10 minutes."
	recipient := to

	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(sender, subject, body, recipient)

	_, _, err := mg.Send(message)

	if err != nil {
		logrus.Warn(err)
		return false
	}
	return true
}

//SendPasswordResetLink ...
func SendPasswordResetLink(to string, link string) bool {
	//mailgun config
	mg := mailgun.NewMailgun(os.Getenv("DOMAIN"), os.Getenv("MAILGUN_API_KEY"))

	sender := ""
	subject := "Reset Your Password"
	body := "Click on this link to reset your password: " + link + " " + "\nValid for 10 minutes."
	recipient := to

	message := mg.NewMessage(sender, subject, body, recipient)

	_, _, err := mg.Send(message)

	if err != nil {
		logrus.Warn(err)
		return false
	}
	return true
}
