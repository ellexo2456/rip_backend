package dsn

import (
	"fmt"
	"os"
)

func FromEnv() string {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return ""
	}

	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
