package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/go-redis/redis/v8"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type MusicService struct {
	redis    *redis.Client
	onBehalf string
}

func InitMusic(redis *utils.MultiRedis, onBehalf string) *MusicService {
	return &MusicService{redis: redis.Get("music"), onBehalf: onBehalf}
}

func (m *MusicService) CleanEmptyNewgrounds() int {
	items, _ := m.redis.SMembers(context.Background(), "nomusic").Result()
	cnt, _ := m.redis.SRem(context.Background(), "nomusic", items).Result()
	return int(cnt)
}

//region Get Music

func (m *MusicService) GetMusicNG(id string) (*structs.MusicResponse, error) {
	return m.GetExMusic("ng", id)
}

func (m *MusicService) GetExMusic(mtype string, id string) (*structs.MusicResponse, error) {
	mus := structs.MusicResponse{}
	if mtype == "ng" && m.redis.SIsMember(context.Background(), "nomusic", id).Val() {
		return nil, errors.New("nomusic")
	}
	if m.redis.Exists(context.Background(), mtype+"::"+id).Val() > 0 {
		v, _ := m.redis.Get(context.Background(), mtype+"::"+id).Result()
		err := json.Unmarshal([]byte(v), &mus)
		if err != nil {
			log.Println(err)
		}
		return &mus, err
	}
	switch mtype {
	case "ng":
		resp, err := http.Get("https://api.fruitspace.one/hmusic/newgrounds?track=" + id + "&onbehalf=" + m.onBehalf)
		if err != nil {
			return nil, err
		}
		rsp, _ := io.ReadAll(resp.Body)
		json.Unmarshal(rsp, &mus)
	case "sc":
		resp, err := http.Get("https://api.fruitspace.one/hmusic/soundcloud?track=https://soundcloud.com/" + id + "&onbehalf=" + m.onBehalf)
		if err != nil {
			return nil, err
		}
		rsp, _ := io.ReadAll(resp.Body)
		json.Unmarshal(rsp, &mus)
	case "vk":
		resp, err := http.Get("https://api.fruitspace.one/hmusic/vk?track=" + id + "&onbehalf=" + m.onBehalf)
		if err != nil {
			return nil, err
		}
		rsp, _ := io.ReadAll(resp.Body)
		json.Unmarshal(rsp, &mus)
	case "yt":
		resp, err := http.Get("https://api.fruitspace.one/hmusic/youtube?track=https://youtube.com/watch?v=" + id + "&onbehalf=" + m.onBehalf)
		if err != nil {
			return nil, err
		}
		rsp, _ := io.ReadAll(resp.Body)
		json.Unmarshal(rsp, &mus)
		mus.Url = "https://cdn2.fruitspace.one/music/yt_" + id + ".mp3"
	case "dz":
		resp, err := http.Get("https://api.fruitspace.one/hmusic/deezer?track=https://deezer.page.link/" + id + "&onbehalf=" + m.onBehalf)
		if err != nil {
			return nil, err
		}
		rsp, _ := io.ReadAll(resp.Body)
		json.Unmarshal(rsp, &mus)
		if mus.Status == "ok" {
			id = strings.Split(mus.Url, "_")[1]
		}
		mus.Url = "https://cdn2.fruitspace.one/music/" + mus.Url + ".mp3"
	}
	if mus.Status == "ok" {
		vjson, _ := json.Marshal(mus)
		m.redis.Set(context.Background(), mtype+"::"+id, string(vjson), 0)
		mus.Url = strings.ReplaceAll(mus.Url, "\\", "")
		return &mus, nil
	}
	if mtype == "ng" {
		m.redis.SAdd(context.Background(), "nomusic", id)
	}
	return nil, errors.New("music not found")
}

func (m *MusicService) TransformHalResource(url string) (*structs.MusicResponse, error) {
	arn := strings.Split(url, ":")
	if len(arn) != 3 {
		return nil, errors.New("invalid url")
	}
	switch arn[1] {
	case "ng":
		if f, _ := regexp.MatchString(`[^0-9]`, arn[2]); f {
			return nil, errors.New("invalid url")
		}
		return m.GetMusicNG(arn[2])
	case "dz":
		if f, _ := regexp.MatchString(`[^0-9]`, arn[2]); f {
			return nil, errors.New("invalid url")
		}
	case "sc":
		if f, _ := regexp.MatchString(`(?i)([a-z\d\-\_])+[\\\\\/]([a-z\d\-\_])+$`, arn[2]); !f {
			return nil, errors.New("invalid url")
		}
	case "yt":
		if f, _ := regexp.MatchString(`(?i)^([a-z\d\-\_])+$`, arn[2]); !f {
			return nil, errors.New("invalid url")
		}
	case "vk":
		if f, _ := regexp.MatchString(`^(\d)+\_(\d)+$`, arn[2]); !f {
			return nil, errors.New("invalid url")
		}
	default:
		return nil, errors.New("invalid url")
	}
	return m.GetExMusic(arn[1], arn[2])
}

//endregion

//region Transforms

func (m *MusicService) GetNG(song string) (string, *structs.MusicResponse, error) {
	song = strings.Split(song, "?")[0]
	re := regexp.MustCompile(`^(((http|https):\/\/|)(www\.|)newgrounds\.com\/audio\/listen\/(\d)+|(\d)+)$`)
	if !re.MatchString(song) {
		return "", nil, errors.New("invalid url")
	}

	//check if song is url
	if f, _ := regexp.MatchString(`[^0-9]`, song); f {
		song = strings.Split(song, "/listen/")[1]
	}
	mr, e := m.GetExMusic("ng", song)
	return song, mr, e
}

func (m *MusicService) GetYT(song string) (string, *structs.MusicResponse, error) {
	re := regexp.MustCompile(`^((http|https):\/\/|)(www\.|m\.|)(youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z\d\-\_]+)([?&].+|)$`)
	if !re.MatchString(song) {
		return "", nil, errors.New("invalid url")
	}
	if strings.Contains(song, "v=") {
		song = strings.Split(song, "v=")[1]
		song = strings.Split(song, "&")[0]
	} else if strings.Contains(song, "youtu.be") {
		v := strings.Split(song, "/")
		song = v[len(v)-1]
		song = strings.Split(song, "?")[0]
	}

	mr, e := m.GetExMusic("yt", song)
	return song, mr, e
}

func (m *MusicService) GetVK(song string) (string, *structs.MusicResponse, error) {
	re := regexp.MustCompile(`^(-|)(\d)+\_(\d)+$`)
	if !re.MatchString(song) {
		return "", nil, errors.New("invalid url")
	}

	mr, e := m.GetExMusic("vk", song)
	return song, mr, e
}

func (m *MusicService) GetDZ(song string) (string, *structs.MusicResponse, error) {
	re := regexp.MustCompile(`^(((http|https):\/\/|)(deezer.page.link/|)([a-zA-Z\d])+)$`)
	if !re.MatchString(song) {
		return "", nil, errors.New("invalid url")
	}
	v := strings.Split(song, "/")
	song = v[len(v)-1]

	mr, e := m.GetExMusic("dz", song)
	return song, mr, e
}

//endregion
