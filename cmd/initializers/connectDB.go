package initializers

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var (
	host     = os.Getenv("DB_HOST")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname   = os.Getenv("DB_NAME")
	port     = os.Getenv("DB_PORT")
)

func ConnectDB() error {
	var err error
	dbConnStr := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=disable TimeZone=Asia/Kolkata"
	DB, err = gorm.Open(postgres.Open(dbConnStr), &gorm.Config{})

	return err
}
