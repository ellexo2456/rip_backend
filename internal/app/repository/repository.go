package repository

import (
	"RIpPeakBack/internal/app/ds"
	"errors"
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

func (r *Repository) GetAlpinistByID(id int) (*ds.Alpinist, error) {
	alpinist := &ds.Alpinist{}

	err := r.db.Preload("Expeditions").First(alpinist, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return alpinist, nil
}

func (r *Repository) GetActiveAlpinists() (*[]ds.Alpinist, error) {
	alpinists := &[]ds.Alpinist{}
	err := r.db.Find(alpinists, "status = ?", "действует").Error

	if err != nil {
		return nil, err
	}

	return alpinists, nil
}

func (r *Repository) FilterByCountry(country string) (*[]ds.Alpinist, error) {
	alpinists := &[]ds.Alpinist{}
	err := r.db.Find(alpinists, "country = ?", country).Error
	if err != nil {
		return nil, err
	}

	return alpinists, nil
}

func (r *Repository) AddAlpinist(alpinist ds.Alpinist) (uint, error) {
	result := r.db.Create(&alpinist)

	if err := result.Error; err != nil {
		return 0, err
	}
	return alpinist.ID, nil
}

func (r *Repository) UpdateAlpinist(alpinist ds.Alpinist) error {
	result := r.db.Save(&alpinist)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddExpedition(expedition ds.Expedition) (uint, error) {
	var existingExpedition ds.Expedition
	result := r.db.Where("status = ?", expedition.Status).First(&existingExpedition)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err := r.db.Create(&expedition).Error
		if err != nil {
			return 0, err
		}
		return expedition.ID, nil
	} else {
		err := r.db.Model(&existingExpedition).Association("Alpinists").Append(expedition.Alpinists)
		if err != nil {
			return 0, err
		}
		return existingExpedition.ID, nil
	}

}

func (r *Repository) GetExpeditionById(id uint) (*ds.Expedition, error) {
	expedition := &ds.Expedition{}

	err := r.db.First(expedition, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (r *Repository) UpdateExpedition(expedition ds.Expedition) error {
	result := r.db.Save(&expedition)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) FilterByStatus(status string, sc ds.SessionContext) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}

	var err error
	if sc.Role == ds.Usr {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "user_id = ? AND status = ?", sc.UserID, status).Error
	} else {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "status = ?", status).Error
	}

	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (r *Repository) FilterByFormedTime(startTime string, endTime string, sc ds.SessionContext) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}

	var err error
	if sc.Role == ds.Usr {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "user_id = ? AND status != 'удалено' AND formed_at BETWEEN ? AND ?", sc.UserID, startTime, endTime).Error
	} else {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "status != 'удалено' AND formed_at BETWEEN ? AND ?", startTime, endTime).Error
	}

	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (r *Repository) FilterByFormedTimeAndStatus(startTime string, endTime string, status string, sc ds.SessionContext) (*[]ds.Expedition, error) {
	expedition := &[]ds.Expedition{}

	var err error
	if sc.Role == ds.Usr {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "user_id = ? AND status = ? AND formed_at BETWEEN ? AND ?", sc.UserID, status, startTime, endTime).Error
	} else {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expedition, "status = ? AND formed_at BETWEEN ? AND ?", status, startTime, endTime).Error
	}
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (r *Repository) GetExpeditions(sc ds.SessionContext) (*[]ds.Expedition, error) {
	expeditions := &[]ds.Expedition{}

	var err error
	if sc.Role == ds.Usr {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expeditions, "status != 'удалено' AND user_id = ?", sc.UserID).Error
	} else {
		err = r.db.Preload("Usr", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, login")
		}).Find(expeditions, "status != 'удалено'").Error
	}

	if err != nil {
		return nil, err
	}

	return expeditions, nil
}

func (r *Repository) UpdateStatus(expedition ds.Expedition) error {
	//result := repository.db.Table("expeditions").Where("id = ?", expedition.ID).Updates(ds.Expedition{Status: expedition.Status, FormedAt: expedition.FormedAt, ClosedAt: expedition.ClosedAt})
	result := r.db.Table("expeditions").Where("id = ?", expedition.ID).Updates(expedition)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetExpeditionByID(expeditionID int) (*ds.Expedition, error) {
	expedition := &ds.Expedition{}

	err := r.db.Preload("Alpinists").First(expedition, "id = ?", expeditionID).Error
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

func (r *Repository) DeleteExpedition(expedition ds.Expedition) error {
	//for _, alpinist := range expedition.Alpinists {
	//	err := repository.db.Model(&expedition).Association("Alpinists").Delete(&alpinist)
	//	if err != nil {
	//		return err
	//	}
	//}

	err := r.db.Updates(ds.Expedition{ID: expedition.ID, Status: expedition.Status, ClosedAt: expedition.ClosedAt}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetDraft(userID int) (ds.Expedition, error) {
	var expedition ds.Expedition
	err := r.db.First(&expedition, "user_id = ? AND status = ?", userID, ds.StatusDraft).Error

	if err != nil {
		return ds.Expedition{}, err
	}

	return expedition, nil
}

func (r *Repository) GetByEmail(email string) (ds.User, error) {
	var u ds.User
	err := r.db.First(&u, "email = ?", email).Error

	if err != nil {
		return ds.User{}, err
	}

	return u, nil
}

func (r *Repository) AddUser(user ds.User) (int, error) {
	result := r.db.Create(&user)

	if err := result.Error; err != nil {
		return 0, err
	}
	return int(user.ID), nil
}
