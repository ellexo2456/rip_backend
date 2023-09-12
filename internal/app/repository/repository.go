package repository

import (
	"RIpPeakBack/internal/app/ds"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
}

func (repository *Repository) GetAlpinistByID(id int) (*ds.Alpinist, error) {
	alpinist := &ds.Alpinist{}

	err := repository.db.First(alpinist, "id = ?", "1").Error // find alpinist with code D42
	if err != nil {
		return nil, err
	}

	return alpinist, nil
}

func (repository *Repository) GetAllAlpinists() ([]*ds.Alpinist, error) {
	var alpinists []*ds.Alpinist
	err := repository.db.Find(alpinists).Error
	if err != nil {
		return nil, err
	}

	return alpinists, nil
}

func (repository *Repository) FilterByCountry(country string) ([]*ds.Alpinist, error) {
	var alpinists []*ds.Alpinist
	err := repository.db.Find(alpinists, "country = ?", country).Error
	if err != nil {
		return nil, err
	}

	return alpinists, nil
}
