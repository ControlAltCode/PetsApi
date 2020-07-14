package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Veterinary struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Name      string    `gorm:"size:255;not null;unique" json:"name"`
	Address   string    `gorm:"size:255;not null;" json:"address"`
	Phone     string    `gorm:"size:128;null;" json:"phone"`
	User      User      `json:"user"`
	UserID    uint32    `sql:"type:int REFERENCES users(id)" json:"user_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (p *Veterinary) Prepare() {
	p.ID = 0
	p.Name = html.EscapeString(strings.TrimSpace(p.Name))
	p.Address = html.EscapeString(strings.TrimSpace(p.Address))
	p.User = User{}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

func (p *Veterinary) Validate() error {

	if p.Name == "" {
		return errors.New("Required Name")
	}
	if p.Address == "" {
		return errors.New("Required Address")
	}
	if p.UserID < 1 {
		return errors.New("Required User")
	}
	return nil
}

func (p *Veterinary) SaveVeterinary(db *gorm.DB) (*Veterinary, error) {
	var err error
	err = db.Debug().Model(&Veterinary{}).Create(&p).Error
	if err != nil {
		return &Veterinary{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.UserID).Take(&p.User).Error
		if err != nil {
			return &Veterinary{}, err
		}
	}
	return p, nil
}

func (p *Veterinary) FindAllVeterinaries(db *gorm.DB) (*[]Veterinary, error) {
	var err error
	posts := []Veterinary{}
	err = db.Debug().Model(&Veterinary{}).Limit(100).Find(&posts).Error
	if err != nil {
		return &[]Veterinary{}, err
	}
	if len(posts) > 0 {
		for i, _ := range posts {
			err := db.Debug().Model(&User{}).Where("id = ?", posts[i].UserID).Take(&posts[i].User).Error
			if err != nil {
				return &[]Veterinary{}, err
			}
		}
	}
	return &posts, nil
}

func (p *Veterinary) FindVeterinaryByID(db *gorm.DB, pid uint64) (*Veterinary, error) {
	var err error
	err = db.Debug().Model(&Veterinary{}).Where("id = ?", pid).Take(&p).Error
	if err != nil {
		return &Veterinary{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.UserID).Take(&p.User).Error
		if err != nil {
			return &Veterinary{}, err
		}
	}
	return p, nil
}

func (p *Veterinary) UpdateAVeterinary(db *gorm.DB) (*Veterinary, error) {

	var err error
	// db = db.Debug().Model(&Veterinary{}).Where("id = ?", pid).Take(&Veterinary{}).UpdateColumns(
	// 	map[string]interface{}{
	// 		"name":      p.Name,
	// 		"address":    p.Address,
	// 		"updated_at": time.Now(),
	// 	},
	// )
	// err = db.Debug().Model(&Veterinary{}).Where("id = ?", pid).Take(&p).Error
	// if err != nil {
	// 	return &Veterinary{}, err
	// }
	// if p.ID != 0 {
	// 	err = db.Debug().Model(&User{}).Where("id = ?", p.UserID).Take(&p.User).Error
	// 	if err != nil {
	// 		return &Veterinary{}, err
	// 	}
	// }
	err = db.Debug().Model(&Veterinary{}).Where("id = ?", p.ID).Updates(Veterinary{Name: p.Name, Address: p.Address, UpdatedAt: time.Now()}).Error
	if err != nil {
		return &Veterinary{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.UserID).Take(&p.User).Error
		if err != nil {
			return &Veterinary{}, err
		}
	}
	return p, nil
}

func (p *Veterinary) DeleteAVeterinary(db *gorm.DB, pid uint64, uid uint32) (int64, error) {

	db = db.Debug().Model(&Veterinary{}).Where("id = ? and user_id = ?", pid, uid).Take(&Veterinary{}).Delete(&Veterinary{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Veterinary not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
