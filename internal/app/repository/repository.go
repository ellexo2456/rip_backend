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

	err := repository.db.Preload("Expeditions").First(alpinist, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return alpinist, nil
}

func (repository *Repository) GetActiveAlpinists() (*[]ds.Alpinist, error) {
	alpinists := &[]ds.Alpinist{}
	err := repository.db.Find(alpinists, "status = ?", "действует").Error

	if err != nil {
		return nil, err
	}

	return alpinists, nil
}

func (repository *Repository) FilterByCountry(country string) (*[]ds.Alpinist, error) {
	alpinists := &[]ds.Alpinist{}
	err := repository.db.Find(alpinists, "country = ?", country).Error
	if err != nil {
		return nil, err
	}

	return alpinists, nil
}

func (repository *Repository) AddAlpinist(alpinist ds.Alpinist) (uint, error) {
	result := repository.db.Create(&alpinist)

	if err := result.Error; err != nil {
		return 0, err
	}
	return alpinist.ID, nil
}

func (repository *Repository) UpdateAlpinist(alpinist ds.Alpinist) error {
	result := repository.db.Save(&alpinist)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (repository *Repository) AddExpedition(expedition ds.Expedition) (uint, error) {
	result := repository.db.Create(&expedition)

	if err := result.Error; err != nil {
		return 0, err
	}
	return expedition.ID, nil
}

func (repository *Repository) GetExpeditionById(id uint) (*ds.Expedition, error) {
	expedition := &ds.Expedition{}

	err := repository.db.First(expedition, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (repository *Repository) UpdateExpedition(expedition ds.Expedition) error {
	result := repository.db.Save(&expedition)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (repository *Repository) FilterByStatus(userID int, status string) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}
	err := repository.db.Find(expedition, "user_id = ? AND status = ?", userID, status).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (repository *Repository) FilterByFormedTime(userID int, startTime string, endTime string) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}
	err := repository.db.Find(expedition, "user_id = ? AND formed_at BETWEEN ? AND ?", userID, startTime, endTime).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (repository *Repository) FilterByFormedTimeAndStatus(userID int, startTime string, endTime string, status string) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}
	err := repository.db.Find(expedition, "user_id = ? AND status = ? AND formed_at BETWEEN ? AND ?", userID, status, startTime, endTime).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (repository *Repository) GetExpeditions(userID int) (*[]ds.Expedition, error) {
	expeditions := &[]ds.Expedition{}

	err := repository.db.Find(expeditions, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	return expeditions, nil
}

func (repository *Repository) UpdateStatus(status string, id uint) error {
	result := repository.db.Table("expeditions").Where("id = ?", id).Update("status", status)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (repository *Repository) GetExpeditionByID(expeditionID int) (*ds.Expedition, error) {
	expedition := &ds.Expedition{}

	err := repository.db.Preload("Alpinists").First(expedition, "id = ?", expeditionID).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (repository *Repository) DeleteExpedition(expedition ds.Expedition) error {
	for _, alpinist := range expedition.Alpinists {
		err := repository.db.Model(&expedition).Association("Alpinists").Delete(&alpinist)
		if err != nil {
			return err
		}
	}

	err := repository.db.Delete(&expedition).Error
	if err != nil {
		return err
	}

	return nil
}
