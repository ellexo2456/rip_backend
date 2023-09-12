package app

import (
	"RIpPeakBack/internal/app/repository"
)

type Application struct {
	repository *repository.Repository
}

func New() Application {
	return Application{}
}
