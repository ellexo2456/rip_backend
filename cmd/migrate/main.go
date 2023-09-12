package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"RIpPeakBack/internal/app/ds"
	"RIpPeakBack/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	//db.Migrator().CreateConstraint(&ds.User{}, "Expeditions")
	//db.Migrator().CreateConstraint(&ds.User{}, "fk_users_expeditions")
	//
	//db.Migrator().CreateConstraint(&ds.Expedition{}, "AlpinistExpeditions")
	//db.Migrator().CreateConstraint(&ds.Expedition{}, "fk_expeditions_alpinist_expeditions")
	//
	//db.Migrator().CreateConstraint(&ds.Alpinist{}, "AlpinistExpeditions")
	//db.Migrator().CreateConstraint(&ds.Alpinist{}, "fk_alpinists_alpinist_expeditions")

	err = db.AutoMigrate(&ds.Alpinist{}, &ds.User{}, &ds.Expedition{})
	if err != nil {
		panic("cant migrate db")
	}
}
