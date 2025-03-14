package postgresql

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
func init() {
    SyncDatabase()
    SeedUsers()
}

func ConnectToDb() {
	var err error
	dsn := os.Getenv("DB")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(("failed to connect db"))
	}
}
