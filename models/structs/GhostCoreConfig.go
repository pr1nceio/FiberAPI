package structs

type ChestConfig struct {
	ChestSmallOrbsMin     int   `json:"ChestSmallOrbsMin"`
	ChestSmallOrbsMax     int   `json:"ChestSmallOrbsMax"`
	ChestSmallDiamondsMin int   `json:"ChestSmallDiamondsMin"`
	ChestSmallDiamondsMax int   `json:"ChestSmallDiamondsMax"`
	ChestSmallShards      []int `json:"ChestSmallShards"`
	ChestSmallKeysMin     int   `json:"ChestSmallKeysMin"`
	ChestSmallKeysMax     int   `json:"ChestSmallKeysMax"`
	ChestSmallWait        int   `json:"ChestSmallWait"`

	ChestBigOrbsMin     int   `json:"ChestBigOrbsMin"`
	ChestBigOrbsMax     int   `json:"ChestBigOrbsMax"`
	ChestBigDiamondsMin int   `json:"ChestBigDiamondsMin"`
	ChestBigDiamondsMax int   `json:"ChestBigDiamondsMax"`
	ChestBigShards      []int `json:"ChestBigShards"`
	ChestBigKeysMin     int   `json:"ChestBigKeysMin"`
	ChestBigKeysMax     int   `json:"ChestBigKeysMax"`
	ChestBigWait        int   `json:"ChestBigWait"`
}

type ServerConfig struct {
	SrvID         string          `json:"-"`
	SrvKey        string          `json:"-"`
	MaxUsers      int             `json:"MaxUsers"`
	MaxLevels     int             `json:"MaxLevels"`
	MaxComments   int             `json:"MaxComments"`
	MaxPosts      int             `json:"MaxPosts"`
	HalMusic      bool            `json:"-"`
	Locked        bool            `json:"-"`
	TopSize       int             `json:"TopSize"`
	EnableModules map[string]bool `json:"EnableModules"`
}

type SecurityConfig struct {
	DisableProtection bool     `json:"DisableProtection"`
	AutoActivate      bool     `json:"AutoActivate"`
	NoLevelLimits     bool     `json:"NoLevelLimits"`
	BannedIPs         []string `json:"BannedIPs"`
}

type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
}

type LogConfig struct {
	LogEnable    bool `json:"LogEnable"`
	LogDB        bool `json:"LogDB"`
	LogEndpoints bool `json:"LogEndpoints"`
	LogRequests  bool `json:"LogRequests"`
}

type GDPSConfig struct {
	DBConfig       DBConfig       `json:"DBConfig"`
	LogConfig      LogConfig      `json:"LogConfig"`
	ChestConfig    ChestConfig    `json:"ChestConfig"`
	ServerConfig   ServerConfig   `json:"ServerConfig"`
	SecurityConfig SecurityConfig `json:"SecurityConfig"`
}
