package db

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/jinzhu/gorm"
)

func HerokuConnection() string {
	regex := regexp.MustCompile("(?i)^postgres://(?:([^:@]+):([^@]*)@)?([^@/:]+):(\\d+)/(.*)$")
	matches := regex.FindStringSubmatch(os.Getenv("DATABASE_URL"))
	if matches == nil {
		fmt.Println("DATABASE_URL variable must look like: postgres://username:password@hostname:port/dbname (not '%v')", os.Getenv("DATABASE_URL"))
		return os.Getenv("DATABASE_URL")
	}

	sslmode := os.Getenv("PGSSL")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", matches[1], matches[2], matches[3], matches[4], matches[5], sslmode)
}

func GormConn() (gorm.DB, error) {
	fmt.Println(HerokuConnection())
	return gorm.Open("postgres", HerokuConnection())
}

func SetupDB() {
	db, err := GormConn()
	if err != nil {
		log.Println("Error migrating database: ", err.Error())
		return
	}

	log.Println("Checking/running DB migrations...")
	db.LogMode(true)
	db.AutoMigrate(&Vacation{})
	db.AutoMigrate(&Assignment{})
	db.AutoMigrate(&Action{})
}
