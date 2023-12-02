package app

import (
	"RIpPeakBack/internal/app/dsn"
	"RIpPeakBack/internal/app/repository"
	"github.com/joho/godotenv"
)

type Application struct {
	repository *repository.Repository
	rr         repository.Redis
}

func New() (Application, error) {
	_ = godotenv.Load()
	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		return Application{}, err
	}

	return Application{repository: repo, rr: repository.NewRedis()}, nil
}
