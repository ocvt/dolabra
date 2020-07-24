package utils

import (
	"os"
)

type Config struct {
	DBName                   string
	ApiUrl                   string
	CookieDomain             string
	EmailLabel               string
	FrontendUrl              string
	GoogleClientSecret       string
	GoogleClientId           string
	GDriveTripsFolderId      string
	GDriveHomePhotosFolderId string
	SmtpFromFirstNameDefault string
	SmtpFromLastNameDefault  string
	SmtpFromEmailDefault     string
	StripePublicKey          string
	StripeSecretKey          string
	StripeWebhookSecret      string
}

func GetConfig() *Config {
	return &Config{
		DBName:                   "dolabra-sqlite",
		ApiUrl:                   os.Getenv("API_URL"),
		CookieDomain:             os.Getenv("COOKIE_DOMAIN"),
		EmailLabel:               os.Getenv("EMAIL_LABEL"),
		FrontendUrl:              os.Getenv("FRONTEND_URL"),
		GoogleClientSecret:       os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleClientId:           os.Getenv("GOOGLE_CLIENT_ID"),
		GDriveTripsFolderId:      os.Getenv("GDRIVE_TRIPS_FOLDER_ID"),
		GDriveHomePhotosFolderId: os.Getenv("GDRIVE_HOME_PHOTOS_FOLDER_ID"),
		SmtpFromFirstNameDefault: os.Getenv("SMTP_FROM_FIRST_NAME_DEFAULT"),
		SmtpFromLastNameDefault:  os.Getenv("SMTP_FROM_LAST_NAME_DEFAULT"),
		SmtpFromEmailDefault:     os.Getenv("SMTP_FROM_EMAIL_DEFAULT"),
		StripePublicKey:          os.Getenv("STRIPE_PUBLIC_KEY"),
		StripeSecretKey:          os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}
