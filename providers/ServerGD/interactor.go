package ServerGD

import (
	"fmt"
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"log"
)

type ServerGDInteractor struct {
	p          *ServerGDProvider
	db         *gorm.DB
	disposable func()
}

func NewServerGDInteractor(p *ServerGDProvider, db *gorm.DB) *ServerGDInteractor {
	return &ServerGDInteractor{p: p, db: db}
}

func NewServerGDInteractorSession(p *ServerGDProvider, srvid string) *ServerGDInteractor {
	db, err := p.mdb.OpenMutated("gdps", srvid)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil
	}
	return &ServerGDInteractor{p: p, db: db,
		disposable: func() {
			p.mdb.DisposeMutated("gdps", srvid)
		}}
}

// Dispose is a fucking miracle that prevents memory leaks and connection overflows.
// If you forgot to dispose connection, your server could easily blow up
func (i *ServerGDInteractor) Dispose() {
	if i.disposable != nil {
		i.disposable()
	}
}

func (i *ServerGDInteractor) GetRoles() []structs.InjectedGDRole {
	var roles []gdps_db.Role
	var uniRoles []structs.InjectedGDRole
	i.p.mdb.UTable(i.db, (&gdps_db.Role{}).TableName()).Find(&roles)
	for _, rl := range roles {
		var users []gdps_db.UserNano
		i.p.mdb.UTable(i.db, (&gdps_db.User{}).TableName()).Where(gdps_db.User{RoleID: rl.ID}).Find(&users)
		uniRoles = append(uniRoles, structs.InjectedGDRole{
			Role:  rl,
			Users: users,
		})
	}

	return uniRoles
}

func (i *ServerGDInteractor) SetRole(role structs.InjectedGDRole) error {
	users := role.Users
	err := i.p.mdb.UTable(i.db, (&gdps_db.Role{}).TableName()).Save(&role).Error
	if err != nil {
		return err
	}
	var uids []int
	for _, u := range users {
		uids = append(uids, u.UID)
	}
	return i.SetRoleUsers(role.ID, uids)
}

func (i *ServerGDInteractor) SetRoleUsers(roleId int, users []int) error {

	// Flush existing privileges
	i.p.mdb.UTable(i.db, (&gdps_db.User{}).TableName()).Where("role_id=?", roleId).Update(gorm.Column(gdps_db.User{}, "RoleID"), 0)
	return i.p.mdb.UTable(i.db, (&gdps_db.User{}).TableName()).Where("uid IN ?", users).Update(gorm.Column(gdps_db.User{}, "RoleID"), roleId).Error
}

func (i *ServerGDInteractor) SearchUsers(query string) []gdps_db.UserNano {
	if len(query) < 3 {
		return nil
	}
	var users []gdps_db.UserNano
	i.p.mdb.UTable(i.db, (&gdps_db.User{}).TableName()).Where("uname LIKE ?", "%"+query+"%").Find(&users)
	return users
}

func (i *ServerGDInteractor) CountActiveUsersLastWeek() int {
	var cnt int64
	i.p.mdb.UTable(i.db, (&gdps_db.User{}).TableName()).Where(fmt.Sprintf("%s>=(CURRENT_DATE - INTERVAL 7 DAY)", gorm.Column(gdps_db.User{}, "UpdatedAt"))).
		Count(&cnt)
	return int(cnt)
}
