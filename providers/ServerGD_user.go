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
	p.mdb.UTable(db, (&gdps_db.User{}).TableName())
	return &ServerGDUser{p: p, db: db, acc: &gdps_db.User{},
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

//
//func (acc *CAccount) PushSettings() {
//	data := map[string]interface{}{
//		"frS": acc.FrS, "cS": acc.CS, "mS": acc.MS,
//		"youtube": acc.Youtube, "twitch": acc.Twitch, "twitter": acc.Twitter}
//	js, _ := json.Marshal(data)
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET settings=? WHERE uid=?", string(js), acc.Uid)
//}
//
//func (acc *CAccount) LoadChests() {
//	var chests string
//	GDPSDB.QueryRow("SELECT chests FROM gdps_"+acc.SrvId+".users WHERE uid=?", acc.Uid).Scan(&chests)
//	var chst map[string]int
//	json.Unmarshal([]byte(chests), &chst)
//	acc.ChestSmallCount = chst["small_count"]
//	acc.ChestBigCount = chst["big_count"]
//	acc.ChestSmallTime = chst["small_time"]
//	acc.ChestBigTime = chst["big_time"]
//}
//
//func (acc *CAccount) PushChests() {
//	data := map[string]int{"small_count": acc.ChestSmallCount, "big_count": acc.ChestBigCount,
//		"small_time": acc.ChestSmallTime, "big_time": acc.ChestBigTime}
//	js, _ := json.Marshal(data)
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET chests=? WHERE uid=?", string(js), acc.Uid)
//}
//
//func (acc *CAccount) PushVessels() {
//	data := map[string]int{"clr_primary": acc.ColorPrimary, "clr_secondary": acc.ColorSecondary, "cube": acc.Cube, "ship": acc.Ship,
//		"ball": acc.Ball, "ufo": acc.Ufo, "wave": acc.Wave, "robot": acc.Robot, "spider": acc.Spider, "swing": acc.Swing,
//		"jetpack": acc.Jetpack, "trace": acc.Trace, "death": acc.Death}
//	js, _ := json.Marshal(data)
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET vessels=?, iconType=? WHERE uid=?", string(js), acc.IconType, acc.Uid)
//}
//func (acc *CAccount) PushStats() {
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET stars=?,diamonds=?,coins=?,ucoins=?,demons=?,cpoints=?,orbs=?,moons=?,special=?,lvlsCompleted=? WHERE uid=?",
//		acc.Stars, acc.Diamonds, acc.Coins, acc.UCoins, acc.Demons, acc.CPoints, acc.Orbs, acc.Moons, acc.Special, acc.LvlsCompleted, acc.Uid)
//}
//
//func (acc *CAccount) LoadAll() {
//	var vessels, settings string
//	GDPSDB.QueryRow("SELECT uid,uname,passhash,gjphash,email,role_id,isBanned,stars,diamonds,coins,ucoins,"+
//		"demons,cpoints,orbs,moons,special,lvlsCompleted,regDate,accessDate,lastIP,gameVer,blacklist,friends_cnt,friendship_ids,"+
//		"iconType,vessels,settings FROM gdps_"+acc.SrvId+".users WHERE uid=?", acc.Uid).Scan(
//		&acc.Uid, &acc.Uname, &acc.Passhash, &acc.GjpHash, &acc.Email, &acc.Role_id, &acc.IsBanned, &acc.Stars, &acc.Diamonds, &acc.Coins,
//		&acc.UCoins, &acc.Demons, &acc.CPoints, &acc.Orbs, &acc.Moons, &acc.Special, &acc.LvlsCompleted, &acc.RegDate, &acc.AccessDate,
//		&acc.LastIP, &acc.GameVer, &acc.Blacklist, &acc.FriendsCount, &acc.FriendshipIds, &acc.IconType, &vessels, &settings)
//	json.Unmarshal([]byte(vessels), acc)
//	var clrs map[string]int
//	json.Unmarshal([]byte(vessels), &clrs)
//	acc.ColorPrimary = clrs["clr_primary"]
//	acc.ColorSecondary = clrs["clr_secondary"]
//	json.Unmarshal([]byte(settings), acc)
//	acc.Blacklist = lib.QuickComma(acc.Blacklist)
//	acc.FriendshipIds = lib.QuickComma(acc.FriendshipIds)
//}
//
//func (acc *CAccount) GetUIDByUname(uname string, autoSave bool) int {
//	var uid int
//	GDPSDB.QueryRow("SELECT uid FROM gdps_"+acc.SrvId+".users WHERE uname=?", uname).Scan(&uid)
//	if uid == 0 {
//		return -1
//	}
//	if autoSave {
//		acc.Uid = uid
//	}
//	return uid
//}
//
//func (acc *CAccount) GetUnameByUID(uid int) string {
//	var uname string
//	GDPSDB.QueryRow("SELECT uname FROM gdps_"+acc.SrvId+".users WHERE uid=?", uid).Scan(&uname)
//	if uname == "" {
//		return "-1"
//	}
//	return uname
//}

func (u *ServerGDUser) UpdateIP(ip string) {
	u.acc.LastIP = ip
	u.db.Where(gdps_db.User{UID: u.acc.UID}).Updates(gdps_db.User{LastIP: ip})
}

//func (acc *CAccount) UpdateGJP2(gjp2 string) {
//	acc.GjpHash = gjp2
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET gjphash=? WHERE uid=?", acc.GjpHash, acc.Uid)
//}
//
//func (acc *CAccount) CountIPs(ip string) int {
//	var cnt int
//	GDPSDB.QueryRow("SELECT count(*) as cnt FROM gdps_"+acc.SrvId+".users WHERE lastIP=?", ip).Scan(&cnt)
//	return cnt
//}
//
//func (acc *CAccount) UpdateBlacklist(action int, uid int) {
//	acc.LoadSocial()
//	blacklist := strings.Split(acc.Blacklist, ",")
//	if action == CBLACKLIST_BLOCK && !slices.Contains(blacklist, strconv.Itoa(uid)) {
//		blacklist = append(blacklist, strconv.Itoa(uid))
//	}
//	if action == CBLACKLIST_UNBLOCK && slices.Contains(blacklist, strconv.Itoa(uid)) {
//		i := slices.Index(blacklist, strconv.Itoa(uid))
//		blacklist = lib.SliceRemove(blacklist, i)
//	}
//	acc.Blacklist = strings.Join(blacklist, ",")
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET blacklist=? WHERE uid=?", acc.Blacklist, acc.Uid)
//}
//
//func (acc *CAccount) UpdateFriendships(action int, uid int) int {
//	acc.LoadSocial()
//	friendships := strings.Split(acc.FriendshipIds, ",")
//	if action == CFRIENDSHIP_ADD && !slices.Contains(friendships, strconv.Itoa(uid)) {
//		acc.FriendsCount++
//		friendships = append(friendships, strconv.Itoa(uid))
//	} else if action == CFRIENDSHIP_REMOVE && slices.Contains(friendships, strconv.Itoa(uid)) {
//		acc.FriendsCount--
//		i := slices.Index(friendships, strconv.Itoa(uid))
//		friendships = lib.SliceRemove(friendships, i)
//	} else {
//		return -1
//	}
//	acc.FriendshipIds = strings.Join(friendships, ",")
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET friends_cnt=?, friendship_ids=? WHERE uid=?", acc.FriendsCount, acc.FriendshipIds, acc.Uid)
//	return 1
//}
//
//func (acc *CAccount) GetShownIcon() int {
//	switch acc.IconType {
//	case 1:
//		return acc.Ship
//	case 2:
//		return acc.Ball
//	case 3:
//		return acc.Ufo
//	case 4:
//		return acc.Wave
//	case 5:
//		return acc.Robot
//	case 6:
//		return acc.Spider
//	case 0:
//	default:
//		return acc.Cube
//	}
//	return acc.Cube
//}
//
//func (acc *CAccount) GetLeaderboardRank() int {
//	var cnt int
//	GDPSDB.QueryRow("SELECT count(*) as cnt FROM gdps_"+acc.SrvId+".users WHERE stars>=? AND isBanned=0", acc.Stars).Scan(&cnt)
//	return cnt
//}
//
//func (acc *CAccount) GetLeaderboard(atype int, grep []string, globalStars int, limit int) []int {
//	var query string
//	switch atype {
//	case CLEADERBOARD_BY_STARS:
//		query = "SELECT uid FROM gdps_" + acc.SrvId + ".users WHERE stars>0 AND isBanned=0 ORDER BY stars DESC, uname ASC LIMIT " + strconv.Itoa(limit)
//	case CLEADERBOARD_BY_CPOINTS:
//		query = "SELECT uid FROM gdps_" + acc.SrvId + ".users WHERE cpoints>0 AND isBanned=0 ORDER BY cpoints DESC, uname ASC LIMIT " + strconv.Itoa(limit)
//	case CLEADERBOARD_GLOBAL:
//		query = "SELECT X.uid as uid,X.stars FROM ((SELECT uid,stars,uname FROM gdps_" + acc.SrvId + ".users WHERE stars>" + strconv.Itoa(globalStars) + " AND isBanned=0 ORDER BY stars ASC LIMIT 50)"
//		query += " UNION (SELECT uid,stars,uname FROM gdps_" + acc.SrvId + ".users WHERE stars<=" + strconv.Itoa(globalStars) + " AND stars>0 AND isBanned=0 ORDER BY stars DESC LIMIT 50)) as X ORDER BY X.stars DESC, X.uname ASC"
//	case CLEADERBOARD_FRIENDS:
//		friends := strings.Join(grep, ",")
//		query = "SELECT uid FROM gdps_" + acc.SrvId + ".users WHERE stars>0 AND isBanned=0 and uid IN (" + friends + ") ORDER BY stars DESC, uname ASC"
//	default:
//		query = "SELECT uid FROM gdps_" + acc.SrvId + ".users WHERE 1=0" //IDK WHY I DID THIS
//	}
//	rows, _ := GDPSDB.Query(query)
//	defer rows.Close()
//
//	var users []int
//	for rows.Next() {
//		var uid int
//		if atype == CLEADERBOARD_GLOBAL {
//			var stars int
//			rows.Scan(&uid, &stars) //Workaround for Globals
//		} else {
//			rows.Scan(&uid)
//		}
//		users = append(users, uid)
//	}
//	return users
//}
//
//func (acc *CAccount) UpdateRole(role_id int) {
//	acc.Role_id = role_id
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET role_id=? WHERE uid=?", role_id, acc.Uid)
//}
//
//func (acc *CAccount) GetRoleObj(fetchPrivs bool) Role {
//	role := Role{}
//	if acc.Role_id == 0 {
//		return role
//	}
//	if fetchPrivs {
//		var privs string
//		GDPSDB.QueryRow("SELECT roleName,commentColor,modLevel,privs FROM gdps_"+acc.SrvId+".roles WHERE id=?", acc.Role_id).Scan(
//			&role.RoleName, &role.CommentColor, &role.ModLevel, &privs)
//		json.Unmarshal([]byte(privs), &role.Privs)
//	} else {
//		GDPSDB.QueryRow("SELECT roleName,commentColor,modLevel FROM gdps_"+acc.SrvId+".roles WHERE id=?", acc.Role_id).Scan(
//			&role.RoleName, &role.CommentColor, &role.ModLevel)
//	}
//	return role
//}
//
//func (acc *CAccount) UpdateAccessTime() {
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET accessDate=?  WHERE uid=?", time.Now().Format("2006-01-02 15:04:05"), acc.Uid)
//}
//
//func (acc *CAccount) BanUser(action int) {
//	var ban int
//	switch action {
//	case CBAN_BAN:
//		ban = 2
//	case CBAN_UNBAN:
//		ban = 0
//	default:
//		ban = 1
//	}
//	acc.IsBanned = ban
//	GDPSDB.Exec("UPDATE gdps_"+acc.SrvId+".users SET isBanned=? WHERE uid=?", ban, acc.Uid)
//}

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
	server.Port = 587
	server.Username = u.p.config["email"]
	server.Password = u.p.config["email_pass"]
	server.Encryption = email.EncryptionSTARTTLS
	client, err := server.Connect()
	if err != nil {
		return err
	}

	msg, err := u.p.assets.ReadFile("assets/GDPSForgotPassword.html")
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

//func (acc *CAccount) Register(uname string, pass string, email string, ip string, autoVerify bool) int {
//	isBanned := "1"
//	if autoVerify {
//		isBanned = "0"
//	}
//	if len(uname) > 16 || !lib.FilterEmail(email) {
//		return -1
//	}
//	if acc.GetUIDByUname(uname, false) != -1 {
//		return -2
//	}
//	var uid int
//	GDPSDB.QueryRow("SELECT uid FROM gdps_"+acc.SrvId+".users WHERE email=?", email).Scan(&uid)
//	if uid != 0 {
//		return -3
//	}
//	//passx := lib.MD5(lib.MD5(pass+"HalogenCore1704")+"ae07") + lib.MD5(pass)[:4]
//	passx := lib.SHA256(lib.SHA512(pass) + "SaltyTruth:sob:")
//
//	rdate := time.Now().Format("2006-01-02 15:04:05")
//	sreq, _ := GDPSDB.Exec(
//		"INSERT INTO gdps_"+acc.SrvId+".users (uname,passhash,gjphash,email,regDate,accessDate,isBanned) VALUES (?,?,?,?,?,?,?)",
//		uname, passx, lib.DoGjp2(pass), email, rdate, rdate, isBanned)
//	vuid, _ := sreq.LastInsertId()
//	acc.Uid = int(vuid)
//	acc.UpdateIP(ip)
//	return 1
//}
