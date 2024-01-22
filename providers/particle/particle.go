package particle

import (
	"context"
	"fmt"
	gorm "github.com/cradio/gormx"
	dbmodels "github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"golang.org/x/exp/slices"
)

var arches = []string{"w64", "w32", "l64", "l32", "l64a", "d64", "d64a"}

type ParticleProvider struct {
	db *gorm.DB
}

func NewParticleProvider(db *gorm.DB) *ParticleProvider {
	db.AutoMigrate(&dbmodels.Particle{})
	db.AutoMigrate(&dbmodels.ParticleUser{})
	return &ParticleProvider{db: db}
}

func (p *ParticleProvider) NewUser() *ParticleUser {
	return &ParticleUser{
		p:    p,
		Data: &dbmodels.ParticleUser{},
	}
}

// SearchParticles searches particles for particle hub, ignoring tags and arches
func (p *ParticleProvider) SearchParticles(query string, arches []string, official bool, sortBy string, page int) (particles []dbmodels.Particle, count int64, err error) {
	if !slices.Contains([]string{"downloads", "likes", "updated_at"}, sortBy) {
		sortBy = "downloads"
	}
	if page < 1 {
		page = 1
	}
	vdb := p.db.WithContext(context.Background())
	vdb = vdb.Model(&dbmodels.Particle{}).Where("is_unlisted = ?", false).Where("is_private = ?", false)
	if len(query) > 2 {
		vdb = vdb.Where("name LIKE ?", "%"+query+"%").Or("author LIKE ?", "%"+query+"%")
	}
	if len(arches) > 0 {
		vdb = vdb.Where("arch IN ?", arches)
	}
	if official {
		vdb = vdb.Where("is_official = ?", true)
	}
	vdb = vdb.Select("name", "author", "uid", "GROUP_CONCAT(DISTINCT arch SEPARATOR ',') as arch", "sum(downloads) as downloads",
		"sum(likes) as likes", "max(is_official) as is_official", "max(updated_at) as updated_at").
		Group("name").Group("author").Group("uid").Order(fmt.Sprintf("%s DESC", sortBy))

	err = vdb.Count(&count).Error
	if err != nil {
		return
	}

	err = vdb.Offset((page - 1) * 50).Limit(50).Find(&particles).Error
	return
}

func (p *ParticleProvider) GetParticle(name string, author string, callerUID int) (particle *structs.ParticleStruct, err error) {
	var particleBranches []dbmodels.Particle
	var particleMeta dbmodels.Particle

	qdb := p.db.WithContext(context.Background()).Where(&dbmodels.Particle{Name: name, Author: author}).
		Where("is_private = ?", false).Or("is_private = ? AND uid = ?", true, callerUID)

	err = qdb.WithContext(context.Background()).Select("id", "max(updated_at) as updated_at", "name", "author", "uid",
		"version", "description", "sum(downloads) as downloads", "max(is_official) as is_official").First(&particleMeta).Error
	if err != nil {
		return
	}
	err = qdb.WithContext(context.Background()).Select("id", "arch", "version", "size").Find(&particleBranches).Error
	if err != nil {
		return
	}

	particle = &structs.ParticleStruct{
		Particle: particleMeta,
	}
	branches := make(map[string][]structs.ParticleBranchItem)
	for _, prt := range particleBranches {
		if _, ok := branches[prt.Version]; !ok {
			branches[prt.Version] = make([]structs.ParticleBranchItem, 0)
		}
		branches[prt.Version] = append(branches[prt.Version], structs.ParticleBranchItem{
			ID:   prt.ID,
			Arch: prt.Arch,
			Size: prt.Size,
		})
	}
	particle.Branches = branches
	return
}

/*
p -> {
	branch1 -> {id, arhes}
	branch2 -> {id, arhes}
}
*/
