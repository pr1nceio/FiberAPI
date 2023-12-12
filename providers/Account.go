package providers

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	email "github.com/xhit/go-simple-mail/v2"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//region AccountProvider

type AccountProvider struct {
	db       *gorm.DB
	redis    *utils.MultiRedis
	keys     map[string]string
	config   map[string]string
	s3config map[string]string
	assets   *embed.FS
}

func NewAccountProvider(dbm *gorm.DB, mredis *utils.MultiRedis) *AccountProvider {
	return &AccountProvider{db: dbm, redis: mredis}
}

func (ap *AccountProvider) WithKeys(keys, config, s3config map[string]string) *AccountProvider {
	ap.keys = keys
	ap.config = config
	ap.s3config = s3config
	return ap
}

func (ap *AccountProvider) WithAssets(assets *embed.FS) *AccountProvider {
	ap.assets = assets
	return ap
}

func (ap *AccountProvider) New() *Account {
	return &Account{
		user: &db.User{},
		p:    ap,
	}
}

func (ap *AccountProvider) GetDiscordIntegrations(onlyClients bool) []string {
	var ids []string
	var users []db.User
	sgd := db.ServerGd{}
	mc := ap.db.Model(db.User{}).Where(fmt.Sprintf("%s!=0", gorm.Column(db.User{}, "DiscordID")))
	if onlyClients {
		// SELECT count(*) from servers_gd WHERE owner_id=uid AND plan>1)
		mc = mc.Where("(?)",
			ap.db.Model(sgd).Select("count(*)").
				Where(fmt.Sprintf("%s=%s", gorm.Column(sgd, "OwnerID"), gorm.Column(db.User{}, "UID"))).
				Where(fmt.Sprintf("%s>1", gorm.Column(sgd, "Plan"))),
		)
	}
	mc.SelectFields(db.User{}, "DiscordID").Find(&users)
	for _, u := range users {
		ids = append(ids, u.DiscordID)
	}
	return ids
}

func (ap *AccountProvider) GetUserCount() int {
	var cnt int64
	ap.db.Model(db.User{}).Count(&cnt)
	return int(cnt)
}

//endregion

// Account is a container for db.User functionality
type Account struct {
	user *db.User
	p    *AccountProvider
}

// calcPassHash calculates password hash
func (a *Account) calcPassHash(uname, password string) string {
	return utils.SHA1(a.p.keys["key_enc"]+password[:6]) + utils.MD5(uname + a.p.keys["key_void"] + password)[1:5]
}

//region Get User

func (a *Account) Data() *db.User {
	return a.user
}

func (a *Account) GetUserByUID(uid int) bool {
	return a.p.db.First(&a.user, uid).Error == nil
}

func (a *Account) GetUserByDiscord(discord_id string) bool {
	return a.p.db.Where(db.User{DiscordID: discord_id}).First(&a.user).Error == nil
}

func (a *Account) GetUserBySession(session string) bool {
	u := db.User{}
	if v, err := a.p.redis.Get("sessions").Get(context.Background(), session).Result(); err == nil {
		u.UID, _ = strconv.Atoi(v)
	} else {
		return false
	}
	return a.p.db.Where(u).First(&a.user).Error == nil
}

func (a *Account) GetUIDByReflink(reflink string) int {
	if len(reflink) == 0 {
		return 0
	}

	u := db.User{}
	a.p.db.Select("uid").Where(&db.User{Reflink: reflink}).First(&u)
	return u.UID
}

//endregion

//region Account Updates

func (a *Account) UpdateNameSurname(name, surname string) error {
	if len(name) < 2 {
		return errors.New("Name is too short |name_shrt")
	}
	if len(name) > 120 {
		return errors.New("Name is too long |name_long")
	}
	if len(surname) < 2 {
		return errors.New("Surname is too short |surname_shrt")
	}
	if len(surname) > 120 {
		return errors.New("Surname is too long |surname_long")
	}
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(name) {
		return errors.New("Invalid name |name")
	}
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(surname) {
		return errors.New("Invalid surname |surname")
	}

	a.user.Name = name
	a.user.Surname = surname

	return a.p.db.Model(a.user).Updates(db.User{Name: name, Surname: surname}).Error
}

func (a *Account) UpdatePassword(pass, newPass string) error {
	if len(pass) < 8 || a.calcPassHash(a.user.Uname, pass) != a.user.PassHash {
		return errors.New("Invalid password |nopwd")
	}
	if len(newPass) < 8 {
		return errors.New("Password is too short |pwd_shrt")
	}
	if len(newPass) > 128 {
		return errors.New("Password is too long |pwd_long")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9~!@#$%^&*()_\-+={}\[\]|\\:;"'<,>.?/]+$`).MatchString(newPass) {
		return errors.New("Invalid password characters |pwd")
	}

	return a.p.db.Model(a.user).Updates(db.User{PassHash: a.calcPassHash(a.user.Uname, newPass)}).Error
}

func (a *Account) UpdateProfilePic(img []byte) error {
	theImg, _, _ := image.Decode(bytes.NewReader(img))
	cropImg := utils.CropSquareImage(theImg)
	var buf bytes.Buffer
	if err := utils.Should(jpeg.Encode(&buf, cropImg, nil)); err != nil {
		return err
	}
	newImg, _ := io.ReadAll(&buf)

	creds := credentials.NewStaticCredentials(a.p.s3config["access_key"], a.p.s3config["secret"], "")
	cfg := aws.NewConfig().WithEndpoint(a.p.s3config["endpoint"]).WithRegion(a.p.s3config["region"]).WithCredentials(creds)
	sess, err := session.NewSession()
	if utils.Should(err) != nil {
		return err
	}
	svc := s3.New(sess, cfg)
	ProfilePic := utils.MD5(strconv.Itoa(int(a.user.UID))) + ".png"
	a.user.ProfilePic = ProfilePic

	params := &s3.PutObjectInput{
		Bucket:        aws.String(a.p.s3config["bucket"]),
		Key:           aws.String("profile_pics/" + ProfilePic),
		Body:          bytes.NewReader(newImg),
		ContentLength: aws.Int64(int64(len(newImg))),
		ContentType:   aws.String("image/jpeg"),
	}
	_, err = svc.PutObject(params)
	if utils.Should(err) != nil {
		return err
	}

	return a.p.db.Model(&a.user).Updates(db.User{ProfilePic: ProfilePic}).Error
}

func (a *Account) ResetProfilePic() {
	a.user.ProfilePic = "default.jpg"
	a.p.db.Model(a.user).Updates(db.User{ProfilePic: a.user.ProfilePic})
}

// UpdateIP updates lastIP and country if needed
func (a *Account) UpdateIP(ip string, fetchGeo bool) {
	a.user.LastIP = ip
	if fetchGeo {
		a.user.Country, a.user.City, a.user.Provider = utils.GetIPRegion(ip, a.p.config["ipinfo_key"])
		a.p.db.Model(a.user).Updates(db.User{
			LastIP:   ip,
			Country:  a.user.Country,
			City:     a.user.City,
			Provider: a.user.Provider,
		})
	} else {
		a.p.db.Model(a.user).Updates(db.User{LastIP: ip})
	}
}

//endregion

// ! TO TRANSFER TO ServerGD as GetCountFor(UID int)
func (a *Account) GetServersCount() map[string]int {
	var cnt int64
	a.p.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: a.user.UID}).Count(&cnt)
	count := map[string]int{
		"gd":  int(cnt),
		"mc":  0,
		"gta": 0,
	}
	return count
}

//region Emails

// DecodeEmailToken decodes UID from token
func (a *Account) DecodeEmailToken(token string) int {
	if len(token) == 0 {
		return 0
	}
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	uid, err := strconv.Atoi(utils.DoXOR(string(decoded), a.p.keys["key_void"]))
	if err != nil {
		return 0
	}
	return uid
}

// EncodeEmailToken encodes UID into email token
func (a *Account) EncodeEmailToken() string {
	return base64.StdEncoding.EncodeToString([]byte(utils.DoXOR(strconv.Itoa(a.user.UID), a.p.keys["key_void"])))
}

// SendEmailVerification actually sends email with formatted html letter
func (a *Account) SendEmailVerification(lang string) error {
	server := email.NewSMTPClient()
	server.Host = a.p.config["email_host"]
	server.Port = 587
	server.Username = a.p.config["email"]
	server.Password = a.p.config["email_pass"]
	server.Encryption = email.EncryptionSTARTTLS
	client, err := server.Connect()
	if err != nil {
		return err
	}

	msg, _ := a.p.assets.ReadFile("assets/EmailConfirm_en.html")
	if lang == "ru" {
		msg, _ = a.p.assets.ReadFile("assets/EmailConfirm_ru.html")
	}
	token := a.EncodeEmailToken()
	msgStr := string(msg)
	msgStr = strings.ReplaceAll(msgStr, "{uname}", a.user.Name+" "+a.user.Surname)
	msgStr = strings.ReplaceAll(msgStr, "{token}", token)

	eml := email.NewMSG()
	eml.SetFrom(a.p.config["email"]).AddTo(a.user.Email).SetSubject("FruitSpace email verification")
	eml.SetBody(email.TextHTML, msgStr)

	return eml.Send(client)
}

// SendEmailRecovery actually sends password recovery email with formatted html letter
func (a *Account) SendEmailRecovery(password string, lang string) error {
	server := email.NewSMTPClient()
	server.Host = a.p.config["email_host"]
	server.Port = 587
	server.Username = a.p.config["email"]
	server.Password = a.p.config["email_pass"]
	server.Encryption = email.EncryptionSTARTTLS
	client, err := server.Connect()
	if err != nil {
		return err
	}

	msg, _ := a.p.assets.ReadFile("assets/ForgotPassword_en.html")
	if lang == "ru" {
		msg, _ = a.p.assets.ReadFile("assets/ForgotPassword_ru.html")
	}
	msgStr := string(msg)
	msgStr = strings.ReplaceAll(msgStr, "{uname}", a.user.Name+" "+a.user.Surname)
	msgStr = strings.ReplaceAll(msgStr, "{password}", password)

	eml := email.NewMSG()
	eml.SetFrom(a.p.config["email"]).AddTo(a.user.Email).SetSubject("Password recovery")
	eml.SetBody(email.TextHTML, msgStr)

	return eml.Send(client)
}

// VerifyEmail sets account state to activated
func (a *Account) VerifyEmail() error {
	if a.user.IsActivated {
		return errors.New("Account is already activated")
	}
	a.user.IsActivated = true
	return a.p.db.Model(a.user).Updates(db.User{IsActivated: true}).Error
}

//endregion

//region Authentication

// NewSession creates new session for Account and puts it in utils.MultiRedis Store
func (a *Account) NewSession(uid int) string {
	// Create new session
	sess := strconv.Itoa(uid) + uuid.New().String()
	a.p.redis.Get("sessions").SetEX(context.Background(), sess, strconv.Itoa(uid), time.Hour*24*30)
	return sess
}

// Register registers user and sends confirmation email
func (a *Account) Register(uname string, name string, surname string, email string, password string, affiliate string, ip string, lang string) error {
	// Register new user
	eml, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("Invalid email |eml")
	}
	email = eml.Address
	if len(uname) < 5 {
		return errors.New("Username is too short |uname_shrt")
	}
	if len(uname) > 32 {
		return errors.New("Username is too long |uname_long")
	}
	if len(name) < 2 {
		return errors.New("Name is too short |name_shrt")
	}
	if len(name) > 120 {
		return errors.New("Name is too long |name_long")
	}
	if len(surname) < 2 {
		return errors.New("Surname is too short |surname_shrt")
	}
	if len(surname) > 120 {
		return errors.New("Surname is too long |surname_long")
	}

	// check if uname is alphanumeric
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]+$`).MatchString(uname) {
		return errors.New("Invalid username |uname")
	}
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(name) {
		return errors.New("Invalid name |name")
	}
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(surname) {
		return errors.New("Invalid surname |surname")
	}

	if len(password) < 8 {
		return errors.New("Password is too short |pwd_shrt")
	}
	if len(password) > 128 {
		return errors.New("Password is too long |pwd_long")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9~!@#$%^&*()_\-+={}\[\]|\\:;"'<,>.?/]+$`).MatchString(password) {
		return errors.New("Invalid password characters |pwd")
	}

	// check if uname is taken
	if a.p.db.Where(db.User{Uname: uname}).First(&db.User{}).Error != gorm.ErrRecordNotFound {
		return errors.New("Username is taken |uname_taken")
	}
	// check if email is taken
	if a.p.db.Where(db.User{Email: email}).First(&db.User{}).Error != gorm.ErrRecordNotFound {
		return errors.New("Email is taken |eml_taken")
	}

	a.user.Uname = uname
	a.user.Name = name
	a.user.Surname = surname
	a.user.Email = email
	a.user.PassHash = a.calcPassHash(uname, password)
	a.user.Affiliate = a.GetUIDByReflink(affiliate)
	a.user.Reflink = utils.MD5(uname + a.p.keys["key_void"])

	if utils.Should(a.p.db.Create(a.user).Error) != nil {
		return errors.New("Unable to register |reg")
	}

	a.UpdateIP(ip, true)

	if utils.Should(a.SendEmailVerification(lang)) != nil {
		return errors.New("Unable to send email verification |ver")
	}
	return nil
}

// Login logs user in
func (a *Account) Login(uname string, password string, ip string) error {
	if a.user.UID != 0 {
		return errors.New("Already logged in |log")
	}

	if a.p.db.Where(db.User{Uname: uname}).First(&a.user).Error != nil {
		return errors.New("Invalid username |nouser")
	}
	if a.user.IsBanned {
		return errors.New("Account is banned |ban")
	}
	if !a.user.IsActivated {
		return errors.New("Account is not activated |act")
	}
	if len(password) < 6 {
		return errors.New("Invalid password |nopwd")
	}

	pass := a.calcPassHash(uname, password)
	if a.user.PassHash != pass {
		return errors.New("Invalid password |nopwd")
	}
	a.UpdateIP(ip, false)
	return nil
}

// RecoverPassword recovers password
func (a *Account) RecoverPassword(email string, lang string) error {
	eml, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("Invalid email |eml")
	}
	email = eml.Address

	if a.p.db.Where(db.User{Email: email}).First(&a.user).Error != nil {
		return errors.New("Invalid email |nouser")
	}
	newPass := utils.GenString(16)

	a.user.PassHash = a.calcPassHash(a.user.Uname, newPass)
	if err := a.p.db.Model(&a.user).Updates(db.User{PassHash: a.user.PassHash}).Error; err != nil {
		return err
	}
	err = a.SendEmailRecovery(newPass, lang)
	return err
}

// CreateTOTP accepts "regen" or valid 2FA Code. Returns secret and base64 image
func (a *Account) CreateTOTP(verify string) (string, string) {
	if a.user.Is2FA {
		if totp.Validate(verify, a.user.TotpSecret) {
			return a.user.TotpSecret, ""
		}
		return "", ""
	}
	if a.user.TotpSecret == "" || verify == "regen" {
		key, _ := totp.Generate(totp.GenerateOpts{
			Issuer:      "FruitSpace",
			AccountName: a.user.Uname,
		})
		a.user.TotpSecret = key.Secret()
		a.p.db.Model(&a.user).Updates(db.User{TotpSecret: a.user.TotpSecret})
		l, _ := key.Image(200, 200)
		var img bytes.Buffer
		png.Encode(&img, l)
		return a.user.TotpSecret, "data:image/png;base64," + base64.StdEncoding.EncodeToString(img.Bytes())
	}
	if totp.Validate(verify, a.user.TotpSecret) {
		a.user.Is2FA = true
		a.p.db.Model(&a.user).Updates(db.User{Is2FA: true})
		return a.user.TotpSecret, ""
	} else {
		return "", ""
	}
}

// AuthDiscord authenticates user by discord code
func (a *Account) AuthDiscord(code, session string) error {
	//https://discord.com/oauth2/authorize?client_id=1119240313605734410&response_type=code&scope=identify%20guilds%20guilds.join&state=<SESSION>
	ds := services.NewDiscordService(fiberapi.DISCORD_CONFIG)
	data, err := ds.AuthByCode(code)
	if err != nil {
		return err
	}
	if len(session) == 0 {
		//Find by discord id
		if a.GetUserByDiscord(data.ClientID) {
			return nil
		}
		return errors.New("No linked user")
	} else {
		if a.GetUserBySession(session) {
			return a.p.db.Model(&a.user).Updates(db.User{
				DiscordID:    data.ClientID,
				DiscordToken: data.Token + ";" + data.RefreshToken,
			}).Error
		}
		// Fall back to new session
		if a.GetUserByDiscord(data.ClientID) {
			return nil
		}
		return errors.New("Unauthorized")
	}
}

func (a *Account) DiscordJoinGuild() error {
	if len(a.user.DiscordToken) == 0 {
		return errors.New("No acc")
	}
	tokens := strings.Split(a.user.DiscordToken, ";")
	ds := services.NewDiscordService(fiberapi.DISCORD_CONFIG)
	t, err := ds.RefreshToken(tokens[1])
	if err != nil {
		return err
	}
	err = ds.JoinGuild(t.Token, a.user.DiscordID)
	if err != nil {
		return err
	}
	return a.p.db.Model(&a.user).Updates(db.User{
		DiscordToken: t.Token + ";" + t.RefreshToken,
	}).Error
}

//endregion
