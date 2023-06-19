package services

import (
	"context"
	"embed"
	"encoding/json"
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"strings"
)

type BuildService struct {
	db          *gorm.DB
	mdb         *utils.MultiSQL
	redis       *redis.Client
	assets      *embed.FS
	s3config    map[string]string
	minioconfig map[string]string
}

func NewBuildService(db *gorm.DB, mdb *utils.MultiSQL, redis *utils.MultiRedis) *BuildService {
	return &BuildService{
		db:    db,
		mdb:   mdb,
		redis: redis.Get("gdps"),
	}
}

func (b *BuildService) WithConfig(s3config, minioconfig map[string]string) *BuildService {
	b.minioconfig = minioconfig
	b.s3config = s3config
	return b
}

func (b *BuildService) WithAssets(assets *embed.FS) *BuildService {
	b.assets = assets
	return b
}

func (b *BuildService) CheckAvail(srvId string) string {
	srvId = srvId[:4]
	if v, _ := b.redis.Exists(context.Background(), srvId).Result(); v == 0 {
		return srvId
	} else {
		return b.CheckAvail(OffsetId(srvId))
	}
}

// InstallServer returns DbPass, SrvKey and Error if any
func (b *BuildService) InstallServer(srvId string, maxUsers int, maxLevels int, maxPosts int, maxComments int) (string, string, error) {
	srvId = srvId[:4]
	Password := utils.GenString(12) + "*"
	SrvKey := utils.GenString(16)
	tx, _ := b.mdb.Raw().BeginTx(context.Background(), nil)
	tx.Exec("DROP USER IF EXISTS halgd_" + srvId + "@localhost")
	tx.Exec("DROP USER IF EXISTS halgd_" + srvId + "@'%'")
	tx.Exec("DROP DATABASE IF EXISTS gdps_" + srvId)
	tx.Exec("CREATE USER halgd_" + srvId + "@localhost IDENTIFIED BY '" + Password + "'")
	tx.Exec("CREATE USER halgd_" + srvId + "@'%' IDENTIFIED BY '" + Password + "'")
	tx.Exec("CREATE DATABASE gdps_" + srvId)
	tx.Exec("GRANT ALL PRIVILEGES ON gdps_" + srvId + ".* TO halgd_" + srvId + "@localhost")
	tx.Exec("GRANT ALL PRIVILEGES ON gdps_" + srvId + ".* TO halgd_" + srvId + "@'%'")
	// !AntiCreate and AntiDrop are experimental
	tx.Exec("REVOKE CREATE,DROP,ALTER ON gdps_" + srvId + ".* FROM halgd_" + srvId + "@localhost")
	tx.Exec("REVOKE CREATE,DROP,ALTER ON gdps_" + srvId + ".* FROM halgd_" + srvId + "@'%'")
	tx.Exec("FLUSH PRIVILEGES")
	if err := tx.Commit(); err != nil {
		return "", "", err
	}
	Config := structs.GenNewGhostConfig(srvId, Password, SrvKey, maxUsers, maxLevels, maxPosts, maxComments)
	if r, _ := b.redis.Exists(context.Background(), srvId).Result(); r != 0 {
		data, _ := b.redis.Get(context.Background(), srvId).Result()
		b.redis.Set(context.Background(), "backup_"+srvId, data, 0)
	}
	jsonC, err := json.Marshal(Config)
	if err != nil {
		return "", "", err
	}
	b.redis.Set(context.Background(), srvId, jsonC, 0)

	itx, err := b.mdb.Raw().BeginTx(context.Background(), nil)
	GDPSDB_SQL, err := b.assets.ReadFile("assets/gdps_database.sql")
	mangle := strings.ReplaceAll(string(GDPSDB_SQL), "CREATE TABLE ", "CREATE TABLE gdps_"+srvId+".")
	itx.Exec(mangle)

	err = itx.Commit()
	return Password, SrvKey, err
}

func (b *BuildService) DeleteServer(srvId, srvName string, alterBucket bool) error {

	S3 := utils.NewS3FS(b.s3config)
	MINIO := utils.NewS3FS(b.minioconfig)

	srvId = srvId[:4]
	// Delete from core
	b.redis.Del(context.Background(), srvId)
	b.redis.Del(context.Background(), "backup_"+srvId)

	// Delete from GDPS DB
	tx, _ := b.mdb.Raw().BeginTx(context.Background(), nil)
	tx.Exec("DROP USER IF EXISTS halgd_" + srvId + "@localhost")
	tx.Exec("DROP USER IF EXISTS halgd_" + srvId + "@'%'")
	tx.Exec("DROP DATABASE IF EXISTS gdps_" + srvId)
	if err := tx.Commit(); err != nil {
		return err
	}

	// Delete from S3
	err := S3.DeleteFolderAsList("gdps_savedata/" + srvId)
	if err != nil {
		log.Println(err)
		err = nil
	}
	err = S3.DeleteFolderAsList("savedata_old/" + srvId)
	if err != nil {
		log.Println(err)
		err = nil
	}
	err = S3.DeleteFile("server_icons/gd_" + srvId + ".png")
	if err != nil {
		log.Println(err)
		err = nil
	}

	log.Println("Deleting " + srvId)

	InstallersS3 := S3
	if alterBucket {
		InstallersS3 = MINIO
		log.Println("Used alternative bucket")
	}
	Folders, err := InstallersS3.ListFolder("gdps_installers/")
	if err != nil {
		log.Println(err)
		InstallersS3.DeleteFile("gdps_installers/" + srvId + "_" + srvName + ".ipa")
		InstallersS3.DeleteFile("gdps_installers/" + srvId + "_" + srvName + ".exe")
		InstallersS3.DeleteFile("gdps_installers/" + srvId + "_" + srvName + ".apk")
	} else {
		for _, v := range Folders {
			if strings.Contains(v, srvId+"_") {
				log.Println("Deleting " + v)
				err = InstallersS3.DeleteFile(v)
				if err != nil {
					log.Println(err)
					err = nil
				}
			}
		}
	}

	// Delete from FS DB
	err = b.db.WhereBinary(db.ServerGd{SrvID: srvId}).Delete(db.ServerGd{}).Error
	err = b.db.Where(db.Queue{Type: "gd"}).WhereBinary(db.Queue{SrvID: srvId}).Delete(db.Queue{}).Error

	// Delete Queue
	return err
}

func (b *BuildService) GetBuildQueue(Worker string) BuilderConfig {
	var job db.Queue
	if b.db.Model(db.Queue{}).Where(db.Queue{Type: "gd"}).Where("worker=''").First(&job).Error != nil {
		return BuilderConfig{}
	}
	var data BuilderConfig
	err := json.Unmarshal([]byte(job.Data), &data)
	if err == nil {
		b.db.Model(job).Updates(db.Queue{Worker: Worker})
	} else {
		log.Println(err)
	}

	return data
}

func (b *BuildService) PushBuildQueue(srvId string, srvName string, icon string, version string,
	bAndroid int, bWindows bool, bIOS bool, bMacOS bool, textures string, region string, alterBucket bool) error {
	srvId = srvId[:4]

	if textures != "default" {
		bAndroid = 2
		go func() {
			S3 := utils.NewS3FS(b.s3config)
			S3.DeleteFile("gdps_textures/" + srvId + ".zip")
			h, _ := http.Get(textures)
			log.Println(S3.PutFileStream("gdps_textures/"+srvId+".zip", h.Body))
		}()
	}

	tPack := TexturePackConfig{
		Enabled:  textures != "default",
		PullFrom: textures,
	}

	bConfig := BuildConfig{
		Android:      bAndroid == 1,
		AndroidUltra: bAndroid == 2,
		Windows:      bWindows,
		IOS:          bIOS,
		MacOS:        bMacOS,
	}

	conf := BuilderConfig{
		Name:        srvName,
		Id:          srvId,
		Region:      region,
		Icon:        icon,
		Version:     version,
		Build:       bConfig,
		TexturePack: tPack,
		AlterBucket: alterBucket,
	}
	data, err := json.Marshal(conf)
	err = b.db.Model(db.Queue{}).Create(db.Queue{Type: "gd", SrvID: srvId, Data: string(data)}).Error
	return err
}

func (b *BuildService) PushClient(srvid string, iType string, file string, alterBucket bool) {
	if alterBucket {
		file = "https://cdn2.fruitspace.one/gdps_installers/" + file
	} else {
		file = "https://cdn.fruitspace.one/gdps_installers/" + file
	}
	cdata := db.ServerGd{}
	switch iType {
	case "a":
		cdata.ClientAndroidURL = file
	case "w":
		cdata.ClientWindowsURL = file
	case "i":
		cdata.ClientIOSURL = file
	case "m":
		cdata.ClientMacOSURL = file
	}
	b.db.Model(db.ServerGd{}).WhereBinary(db.ServerGd{SrvID: srvid}).Updates(cdata)
}

func (b *BuildService) NotifyDone(ltype string, srvid string) {
	b.db.WhereBinary(db.Queue{SrvID: srvid}).Where(db.Queue{Type: ltype}).Delete(db.Queue{})
}

func OffsetId(srvId string) string {
	sid := []byte(srvId)
	switch sid[3] {
	case 57:
		sid[3] = 'A'
	case 90:
		sid[3] = 'a'
	case 122:
		sid[3] = '0'
		switch sid[2] {
		case 57:
			sid[2] = 'A'
		case 90:
			sid[2] = 'a'
		case 122:
			sid[2] = '0'
			switch sid[1] {
			case 57:
				sid[1] = 'A'
			case 90:
				sid[1] = 'a'
			case 122:
				sid[1] = '0'
				switch sid[0] {
				case 57:
					sid[0] = 'A'
				case 90:
					sid[0] = 'a'
				case 122:
					return ""
				default:
					sid[0]++
				}
			default:
				sid[1]++
			}
		default:
			sid[2]++
		}
	default:
		sid[3]++
	}
	return string(sid)
}

type BuildConfig struct {
	Android      bool `json:"android"`
	AndroidUltra bool `json:"android_ultra"`
	Windows      bool `json:"windows"`
	IOS          bool `json:"ios"`
	MacOS        bool `json:"macos"`
}

type TexturePackConfig struct {
	Enabled  bool   `json:"enabled"`
	PullFrom string `json:"pullfrom"`
}

type BuilderConfig struct {
	Name        string            `json:"name"`
	Id          string            `json:"id"`
	Region      string            `json:"region"`
	Icon        string            `json:"icon"`
	Version     string            `json:"version"` //2.1 2.2 1.9 1.0
	Build       BuildConfig       `json:"build"`
	TexturePack TexturePackConfig `json:"texturepack"`
	AlterBucket bool              `json:"alterbucket"`
}
