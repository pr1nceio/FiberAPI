package structs

type GDSettings struct {
	SrvId       string `json:"id"`
	Description struct {
		Text    string `json:"text"`
		Align   int    `json:"align"`
		Discord string `json:"discord"`
		Vk      string `json:"vk"`
	} `json:"description"`
	SpaceMusic bool `json:"spaceMusic"`
	TopSize    int  `json:"topSize"`
	Security   struct {
		Enabled      bool `json:"enabled"`
		AutoActivate bool `json:"autoActivate"`
		LevelLimit   bool `json:"levelLimit"`
	} `json:"security"`
	Modules map[string]bool `json:"modules"`
}

type BuildLabSettings struct {
	SrvName string `json:"srvname"`
	Version string `json:"version"`

	Windows bool `json:"windows"`
	Android bool `json:"android"`
	IOS     bool `json:"ios"`
	MacOS   bool `json:"macos"`

	Icon     string `json:"icon"`
	Textures string `json:"textures"`
}

type LogEntry struct {
	Id       int64             `json:"id" db:"id"`
	Date     string            `json:"date" db:"date"`
	Uid      int64             `json:"uid" db:"uid"`
	Type     int64             `json:"type" db:"type"`
	TargetID int64             `json:"target_id" db:"target_id"`
	IsMod    bool              `json:"isMod" db:"isMod"`
	Data     map[string]string `json:"data" db:"-"`

	Mdata string `json:"-" db:"data"`
	Udata string `json:"-" db:"udata"`
}
