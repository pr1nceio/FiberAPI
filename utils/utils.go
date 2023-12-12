package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/oliamb/cutter"
	"golang.org/x/exp/slices"
	"image"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

var Loc, _ = time.LoadLocation("Europe/Moscow")

func sliceRemove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// GetIPRegion returns country, city, provider
func GetIPRegion(ip string, token string) (string, string, string) {
	r, err := http.Get(fmt.Sprintf("https://ipinfo.io/%s/json?token=%s", ip, token))
	if err != nil {
		return "Unknown", "Unknown", "Unknown"
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "Unknown", "Unknown", "Unknown"
	}
	via := struct {
		Country string `json:"country"`
		City    string `json:"city"`
		Org     string `json:"org"`
	}{}
	err = json.Unmarshal(body, &via)
	if err != nil {
		return "Unknown", "Unknown", "Unknown"
	}
	return via.Country, via.City, via.Org
}

func MD5(str string) string {
	md := md5.New()
	md.Write([]byte(str))
	return fmt.Sprintf("%x", md.Sum(nil))
}

func SHA256(str string) string {
	sha := sha256.New()
	sha.Write([]byte(str))
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func SHA512(str string) string {
	sha := sha512.New()
	sha.Write([]byte(str))
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func SHA1(str string) string {
	sha := sha1.New()
	sha.Write([]byte(str))
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func GetEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func GetKVEnv(key string) map[string]string {
	val := os.Getenv(key)
	if val == "" {
		return map[string]string{}
	}
	kv := map[string]string{}
	for _, v := range strings.Split(val, ",") {
		kv[strings.SplitN(v, "=", 2)[0]] = strings.SplitN(v, "=", 2)[1]
	}
	return kv
}

func ReadPost(req *http.Request) url.Values {
	if req.Body == nil {
		return url.Values{}
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err.Error())
		return url.Values{}
	}
	if len(body) == 0 || strings.Count(string(body), "=") == 0 {
		return url.Values{}
	}
	vals := make(url.Values)
	pairs := strings.Split(string(body), "&")
	for _, val := range pairs {
		if !strings.Contains(val, "=") {
			continue
		}
		m := strings.SplitN(val, "=", 2)
		//fmt.Println(m)
		rval, _ := url.QueryUnescape(m[1])
		rkey, _ := url.QueryUnescape(m[0])
		vals[rkey] = append(vals[rkey], rval)
	}
	return vals
}

func DoXOR(text string, key string) (output string) {
	for i := 0; i < len(text); i++ {
		output += string(text[i] ^ key[i%len(key)])
	}
	return output
}

func VerifyCaptcha(captcha string, secret string) bool {
	r, err := http.Post("https://hcaptcha.com/siteverify", "application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("secret=%s&response=%s", secret, captcha)))
	if err != nil {
		return false
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	res := struct {
		Success bool `json:"success"`
	}{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}
	return res.Success
}

func VerifyFCaptcha(captcha string, secret string) bool {
	//URLENCODE
	data := struct {
		Secret   string `json:"secret"`
		Solution string `json:"solution"`
	}{
		Secret:   secret,
		Solution: captcha,
	}
	d, _ := json.Marshal(data)
	r, err := http.Post("https://api.friendlycaptcha.com/api/v1/siteverify", "application/json", bytes.NewBuffer(d))
	log.Println(string(d), err)
	if err != nil {
		return false
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	log.Println(string(body))
	if err != nil {
		return false
	}
	res := struct {
		Success bool `json:"success"`
	}{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}
	return res.Success
}

func CropSquareImage(img image.Image) image.Image {
	b := img.Bounds()
	w := b.Max.X - b.Min.X
	h := b.Max.Y - b.Min.Y
	sz := int(math.Min(float64(w), float64(h)))
	newImg, _ := cutter.Crop(img, cutter.Config{
		Width:  sz,
		Height: sz,
		Mode:   cutter.Centered,
	})
	return newImg
}

func SendMessageDiscord(text string) {
	b, _ := json.Marshal(map[string]string{
		"content": text,
	})

	content := bytes.NewReader(b)

	http.Post("https://discord.com/api/webhooks/1133486914394144768/YDbd9c1FrS42bNiDRMrTctyWBa0xWASWokCKiz2eAL5SfZkyzMNSd45wUJvY_xA-KrlQ",
		"application/json", content)
}

func GenString(length int) string {
	var str string
	alpha := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < length; i++ {
		str += string(alpha[rand.Intn(len(alpha))])
	}
	return str
}

func Should(err error) error {
	if err != nil {
		log.Println(err)
		log.Output(2, fmt.Sprintln(err))
		sentry.CaptureException(err)
	}
	return err
}

func HideField(model interface{}, field string) []string {
	var fields []string
	m := reflect.TypeOf(model)
	for i := 0; i < m.NumField(); i++ {
		f := m.Field(i)
		val := f.Tag.Get("gorm")
		keys := strings.Split(val, ";")
		for _, key := range keys {
			kv := strings.Split(key, ":")
			if kv[0] == "column" && len(kv) > 1 {
				if f.Name != field {
					fields = append(fields, kv[1])
				}
			}
		}
	}

	return fields
}

func FilterEmail(email string) bool {
	semail := strings.Split(email, "@")
	if len(semail) != 2 {
		return false
	}
	AllowedEmailProviders := []string{
		"yandex.ru",
		"ya.ru",
		"mail.ru",
		"gmail.com",
		"aol.com",
		"rambler.ru",
		"bk.ru",
		"vk.com",
	}
	if !slices.Contains(AllowedEmailProviders, strings.ToLower(semail[1])) {
		return false
	}
	return true
}
