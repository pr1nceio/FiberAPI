package particle

import "fmt"

type Manifest struct {
	Name        string            `json:"name"`
	Author      string            `json:"author"`
	Note        string            `json:"note"`
	Block       string            `json:"block"`
	Server      string            `json:"server,omitempty"`
	LayerServer string            `json:"block_server,omitempty"`
	Meta        map[string]string `json:"meta"`
	Recipe      Recipe            `json:"recipe"`
}

type Recipe struct {
	Base    string   `json:"base"`
	Apply   []string `json:"apply"`
	Engines []string `json:"engines"`
	Run     []string `json:"run"`
}

func QuickGDPSRecipe(srvid string, version string, textureBase string) Manifest {
	base := version
	patcher := "2.1"
	if version == "2.2" {
		base = "2.204"
		patcher = "2.2"
	}
	textureBase = fmt.Sprintf("m41den/gdps_windows@%s", base)

	return Manifest{
		Name:   fmt.Sprintf("%s@%s", srvid, version),
		Author: "ghost",
		Note:   "",
		Block:  "",
		Meta: map[string]string{
			"FS_GDPS":     srvid,
			"FS_GDPS_VER": version,
		},
		Recipe: Recipe{
			Base: textureBase,
			Apply: []string{
				fmt.Sprintf("m41den/gdps_patcher@%s", patcher),
			},
			Engines: []string{},
			Run:     []string{},
		},
	}
}
