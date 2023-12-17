package providers

import (
	"errors"
	"fmt"
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/utils"
	email "github.com/xhit/go-simple-mail/v2"
	"log"
	"regexp"
	"strings"
)

type ServerGDUser struct {
	p          *ServerGDProvider
	db         *gorm.DB
	acc        *gdps_db.User
	disposable func()
}

func NewServerGDUser(p *ServerGDProvider, db *gorm.DB) *ServerGDUser {
	return &ServerGDUser{p: p, db: db, acc: &gdps_db.User{}}
}

func NewServerGDUserSession(p *ServerGDProvider, srvid string) *ServerGDUser {
	db, err := p.mdb.OpenMutated("gdps", srvid)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil
	}
	return &ServerGDUser{p: p, db: p.mdb.UTable(db, (&gdps_db.User{}).TableName()),
		acc: &gdps_db.User{},
		disposable: func() {
			p.mdb.DisposeMutated("gdps", srvid)
		}}
}

// Dispose is a fucking miracle that prevents memory leaks and connection overflows.
// If you forgot to dispose connection, your server could easily blow up
func (u *ServerGDUser) Dispose() {
	if u.disposable != nil {
		u.disposable()
	}
}

func (u *ServerGDUser) Data() *gdps_db.User {
	return u.acc
}

func (u *ServerGDUser) CountUsers() int {
	var cnt int64
	u.db.Count(&cnt)
	return int(cnt)
}

func (u *ServerGDUser) Exists(uid int) bool {
	var cnt int64
	u.db.Where(gdps_db.User{UID: uid}).Count(&cnt)
	return cnt > 0
}

func (u *ServerGDUser) GetUserByUID(uid int) bool {
	return u.db.First(&u.acc, uid).Error == nil
}
func (u *ServerGDUser) GetUserByUname(uname string) bool {
	return u.db.Where(gdps_db.User{Uname: uname}).First(&u.acc).Error == nil
}
func (u *ServerGDUser) GetUserByEmail(email string) bool {
	return u.db.Where(gdps_db.User{Email: email}).First(&u.acc).Error == nil
}

func (u *ServerGDUser) UpdateIP(ip string) {
	u.acc.LastIP = ip
	u.db.Where(gdps_db.User{UID: u.acc.UID}).Updates(gdps_db.User{LastIP: ip})
}

func (u *ServerGDUser) changePassword(passhash string) {
	u.db.Where(gdps_db.User{UID: u.acc.UID}).Updates(gdps_db.User{Passhash: passhash})
}

func (u *ServerGDUser) UserChangePassword(pass string) error {
	if len(pass) < 5 || len(pass) > 32 {
		return errors.New("Password is too short or too long |pwd_shrt")
	}
	pass = utils.SHA256(utils.SHA512(pass) + "SaltyTruth:sob:")
	u.changePassword(pass)
	return nil
}

func (u *ServerGDUser) UserChangeEmail(email string) error {
	if !utils.FilterEmail(email) {
		return errors.New("Invalid email |email")
	}
	u.db.Where(gdps_db.User{UID: u.acc.UID}).Updates(gdps_db.User{Email: email})
	return nil
}

func (u *ServerGDUser) UserChangeUsername(uname string) error {
	if len(uname) > 16 {
		return errors.New("Username is too long |uname_long")
	}
	if len(uname) < 4 {
		return errors.New("Username is too short |uname_shrt")
	}
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]+$`).MatchString(uname) {
		return errors.New("Invalid username |uname")
	}
	u.db.Where(gdps_db.User{UID: u.acc.UID}).Updates(gdps_db.User{Uname: uname})
	return nil
}

func (u *ServerGDUser) UserForgotPasswordSendEmail(srvid string) error {
	server := email.NewSMTPClient()
	server.Host = u.p.config["email_host"]
	server.Port = 25 //587
	server.Username = u.p.config["email"]
	server.Password = u.p.config["email_pass"]
	server.Encryption = email.EncryptionNone //email.EncryptionSTARTTLS
	client, err := server.Connect()
	if err != nil {
		return err
	}

	msg, _ := u.p.assets.ReadFile("assets/GDPSForgotPassword.html")
	msgStr := string(msg)
	msgStr = strings.ReplaceAll(msgStr, "{uname}", u.acc.Uname)
	token := fmt.Sprintf("%d:%s", u.acc.UID, u.acc.Passhash)
	msgStr = strings.ReplaceAll(msgStr, "{url}", fmt.Sprintf("https://gofruit.space/gdps/%s/recover?token=%s", srvid, token))
	msgStr = strings.ReplaceAll(msgStr, "{srvid}", srvid)

	eml := email.NewMSG()
	eml.SetFrom(u.p.config["email"]).AddTo(u.acc.Email).SetSubject("Password recovery")
	eml.SetBody(email.TextHTML, msgStr)

	return eml.Send(client)
}

func (u *ServerGDUser) LogIn(uname string, pass string, ip string, uid int, rawhash bool) int {
	if uid == 0 {
		u.GetUserByUname(uname)
		uid = u.acc.UID
	} else {
		if !u.GetUserByUID(uid) {
			return -1
		}
	}
	if uid > 0 {
		if u.acc.IsBanned > 1 {
			return -12
		}

		passx := utils.SHA256(utils.SHA512(pass) + "SaltyTruth:sob:")
		if len(u.acc.Passhash) == 36 {
			u.changePassword(passx)
			passx = utils.MD5(utils.MD5(pass+"HalogenCore1704")+"ae07") + utils.MD5(pass)[:4]
		}

		if rawhash {
			passx = pass
		}
		if u.acc.Passhash == passx {
			u.UpdateIP(ip)
			u.db.Where(gdps_db.User{UID: u.acc.UID}).UpdateColumn(gorm.Column(gdps_db.User{}, "IsBanned"), 0)
			return uid
		}
	}
	return -1
}
