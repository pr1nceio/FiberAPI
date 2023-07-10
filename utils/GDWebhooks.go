package utils

import (
	"fmt"
	"github.com/aiomonitors/godiscord"
	"strings"
)

var gdpsbot_diffs = map[string]string{
	"unrated":                "<:unrated:949684835818041375>",
	"auto":                   "<:auto:949684856567259197>",
	"auto-featured":          "<:autofeatured:949684833284661348>",
	"auto-epic":              "<:autoepic:949684833259487232>",
	"easy":                   "<:easy:949684834056421396>",
	"easy-featured":          "<:easyfeatured:949684834320678962>",
	"easy-epic":              "<:easyepic:949684834777849887>",
	"normal":                 "<:normal:949684834069000263>",
	"normal-featured":        "<:normalfeatured:949684834240975008>",
	"normal-epic":            "<:normalepic:949684834253570148>",
	"hard":                   "<:hard:949684834194821170>",
	"hard-featured":          "<:hardfeatured:949684834022850591>",
	"hard-epic":              "<:hardepic:949684834475835454>",
	"harder":                 "<:harder:949684834333237309>",
	"harder-featured":        "<:harderfeatured:949684834178068580>",
	"harder-epic":            "<:harderepic:949684834421313646>",
	"insane":                 "<:insane:949684834064797758>",
	"insane-featured":        "<:insanefeatured:949684834073182209>",
	"insane-epic":            "<:insaneepic:949684834324873277>",
	"demon-easy":             "<:demoneasy:949684833498562680>",
	"demon-easy-featured":    "<:demoneasyfeatured:949684833104318525>",
	"demon-easy-epic":        "<:demoneasyepic:949684833636999179>",
	"demon-medium":           "<:demonmedium:949684833825718293>",
	"demon-medium-featured":  "<:demonmediumfeatured:949684834199044136>",
	"demon-medium-epic":      "<:demonmediumepic:949684834127736852>",
	"demon-hard":             "<:demonhard:949684833922199573>",
	"demon-hard-featured":    "<:demonhardfeatured:949684834014466059>",
	"demon-hard-epic":        "<:demonhardepic:949684833804775436>",
	"demon-insane":           "<:demoninsane:949684834387763250>",
	"demon-insane-featured":  "<:demoninsanefeatured:949684833972527184>",
	"demon-insane-epic":      "<:demoninsaneepic:949684834534559784>",
	"demon-extreme":          "<:demonextreme:949684833532141581>",
	"demon-extreme-featured": "<:demonextremefeatured:949684833347567638>",
	"demon-extreme-epic":     "<:demonextremeepic:949684833959944282>",
}

var gdpsbot_emojis = map[string]string{
	"star":      "<:gdstar:949684834203238481>",
	"coin":      "<:silvercoin:949684834119327784>",
	"downloads": "<:downloads:949684940302348308>",
	"like":      "<:gdlike:949684747725045820>",
	"time":      "<:length:949685919131258991>",
}

func timeToStr(time string) string {
	switch time {
	case "0":
		return "Tiny"
	case "1":
		return "Short"
	case "2":
		return "Medium"
	case "3":
		return "Long"
	case "4":
		return "XL"
	default:
		return "Unknown"
	}
}

func rgbToHex(rgb string) string {
	rgbx := strings.Split(rgb, ",")
	if len(rgbx) != 3 {
		return "#fffff"
	}
	return fmt.Sprintf("#%02x%02x%02x", rgbx[0], rgbx[1], rgbx[2])
}

func musicUploaded(music map[string]string) godiscord.Embed {
	embed := godiscord.NewEmbed("Загружен новый трек", "", "")
	embed.SetColor("#0d6efd")
	embed.SetThumbnail("https://cdn.fruitspace.one/assets/bot_icons/disc.png")
	embed.AddField(
		fmt.Sprintf("**%s** - *%s*", music["name"], music["artist"]),
		fmt.Sprintf("ID: **%s**", music["id"]),
		false,
	)
	embed.SetFooter("Добавил "+music["nickname"], "")
	return embed
}

func levelRated(level map[string]string) godiscord.Embed {
	embed := godiscord.NewEmbed("Оценен уровень!", "", "")
	embed.SetColor("#d4af37")
	embed.SetThumbnail(fmt.Sprintf("https://cdn.fruitspace.one/assets/bot_icons/lvl/%s.png", level["diff"]))
	embed.AddField(level["name"], fmt.Sprintf("by *%s*", level["builder"]), false)
	embed.AddField(
		fmt.Sprintf("ID: %s ⠀%s %s", level["id"], gdpsbot_emojis["star"], level["stars"]),
		fmt.Sprintf("%s %s⠀|⠀%s %s | %s %s",
			gdpsbot_emojis["downloads"], level["downloads"], gdpsbot_emojis["like"], level["likes"], gdpsbot_emojis["time"], timeToStr(level["len"])),
		false,
	)
	embed.SetFooter(fmt.Sprintf("Оценил %s", level["rateuser"]), "")
	return embed
}

func levelUploaded(level map[string]string) godiscord.Embed {
	if level["desc"] == "" {
		level["desc"] = "-"
	}
	if level["name"] == "" {
		level["name"] = "-"
	}
	embed := godiscord.NewEmbed("Новый уровень!", "", "")
	embed.SetColor("#cacad0")
	embed.SetThumbnail("https://cdn.fruitspace.one/assets/bot_icons/lvl/unrated.png")
	embed.AddField(level["name"], fmt.Sprintf("by *%s*", level["builder"]), false)
	embed.AddField(fmt.Sprintf(" ID: %s", level["id"]), fmt.Sprintf("Описание: `%s`", level["desc"]), true)
	return embed
}

func newUser(nickname string) godiscord.Embed {
	embed := godiscord.NewEmbed("", "", "")
	embed.SetColor("#0d6efd")
	embed.AddField("🎉 Новый игрок!", fmt.Sprintf("Рады видеть тебя, **%s**", nickname), true)
	return embed
}

func GetEmbed(xtype string, data map[string]string) godiscord.Embed {
	switch xtype {
	case "rate":
		return levelRated(data)
	case "newlevel":
		return levelUploaded(data)
	case "newuser":
		return newUser(data["nickname"])
	case "newmusic":
		return musicUploaded(data)
	}
	return godiscord.NewEmbed("", "", "")
}
