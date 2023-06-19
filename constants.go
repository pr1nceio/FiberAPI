package fiberapi

import (
	"embed"
	"github.com/fruitspace/FiberAPI/utils"
)

var (
	KEYS           = utils.GetKVEnv("KEYS")           //key_enc=,key_void=
	CONFIG         = utils.GetKVEnv("CONFIG")         //ipinfo_key=,email=,email_pass=,email_host=,hCaptchaToken=
	DISCORD_CONFIG = utils.GetKVEnv("DISCORD_CONFIG") //appid=,secret=,url=

	S3_CONFIG    = utils.GetKVEnv("S3_CONFIG") //access_key=,secret=,region=,bucket=,endpoint=,cdn=
	MINIO_CONFIG = utils.GetKVEnv("MINIO_CONFIG")

	DB_HOST = utils.GetEnv("DB_HOST", "localhost:3306")
	DB_NAME = utils.GetEnv("DB_NAME", "default_db")
	DB_USER = utils.GetEnv("DB_USER", "root")
	DB_PASS = utils.GetEnv("DB_PASS", "root")

	GDPSDB_HOST = utils.GetEnv("GDPSDB_HOST", "localhost:3307")
	GDPSDB_USER = utils.GetEnv("GDPSDB_USER", "root")
	GDPSDB_PASS = utils.GetEnv("GDPSDB_PASS", "root")

	REDIS_HOST = utils.GetEnv("REDIS_HOST", "localhost:6379")
	REDIS_PASS = utils.GetEnv("REDIS_PASS", "")

	PAYMENTS_HOST = utils.GetEnv("PAYMENTS_HOST", "http://localhost:8000")

	ADDR = utils.GetEnv("ADDR", ":8080")
)

//go:embed assets
var AssetsDir embed.FS

var ValidImageTypes = []string{
	"image/jpeg",
	"image/png",
}

var ValidMerchants = []string{"qiwi", "enot", "yookassa"}
