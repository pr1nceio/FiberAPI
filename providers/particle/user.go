package particle

import (
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/google/uuid"
)

type ParticleUser struct {
	p    *ParticleProvider
	Data *db.ParticleUser
}

func (p *ParticleUser) GetByUID(uid int) error {
	return p.p.db.Model(&db.ParticleUser{}).First(&p.Data, uint(uid)).Error
}

func (p *ParticleUser) RegisterFromUser(user *db.User) error {
	p.Data.Username = user.Uname
	p.Data.MaxAllowedSize = 512 * 1024 * 1024
	p.Data.Token = "p_" + uuid.NewString()
	p.Data.ID = uint(user.UID)
	return p.p.db.Model(&db.ParticleUser{}).Create(&p.Data).Error
}

func (p *ParticleUser) CalculateUsedSize() (size uint, err error) {
	err = p.p.db.Model(db.Particle{}).Where(db.Particle{UID: p.Data.ID}).Select("sum(size)").Scan(&size).Error
	return
}
