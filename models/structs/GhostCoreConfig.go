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
	SrvID         string                           `json:"SrvID"`
	SrvKey        string                           `json:"SrvKey"`
	MaxUsers      int                              `json:"MaxUsers"`
	MaxLevels     int                              `json:"MaxLevels"`
	MaxComments   int                              `json:"MaxComments"`
	MaxPosts      int                              `json:"MaxPosts"`
	HalMusic      bool                             `json:"HalMusic"`
	Locked        bool                             `json:"Locked"`
	TopSize       int                              `json:"TopSize"`
	EnableModules map[string]bool                  `json:"EnableModules"`
	Modules       map[string](map[string]([]byte)) `json:"Modules"` //str_module: {str_key:b_value}
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

func GenNewGhostConfig(srvId string, DBPass string, SrvKey string, MaxUsers int, MaxLevels int, MaxPosts int, MaxComments int) GDPSConfig {
	return GDPSConfig{
		DBConfig: DBConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "halgd_" + srvId,
			Password: DBPass,
			DBName:   "gdps_" + srvId,
		},
		LogConfig: LogConfig{
			LogEnable:    true,
			LogDB:        false,
			LogEndpoints: false,
			LogRequests:  false,
		},
		ChestConfig: ChestConfig{
			ChestSmallOrbsMin:     200,
			ChestSmallOrbsMax:     400,
			ChestSmallDiamondsMin: 2,
			ChestSmallDiamondsMax: 10,
			ChestSmallShards:      []int{1, 2, 3, 4, 5, 6},
			ChestSmallKeysMin:     1,
			ChestSmallKeysMax:     6,
			ChestSmallWait:        3600,

			ChestBigOrbsMin:     2000,
			ChestBigOrbsMax:     4000,
			ChestBigDiamondsMin: 20,
			ChestBigDiamondsMax: 100,
			ChestBigShards:      []int{1, 2, 3, 4, 5, 6},
			ChestBigKeysMin:     1,
			ChestBigKeysMax:     6,
			ChestBigWait:        14400,
		},
		ServerConfig: ServerConfig{
			SrvID:         srvId,
			SrvKey:        SrvKey,
			MaxUsers:      MaxUsers,
			MaxLevels:     MaxLevels,
			MaxComments:   MaxComments,
			MaxPosts:      MaxPosts,
			HalMusic:      false,
			Locked:        false,
			TopSize:       100,
			EnableModules: map[string]bool{},
			Modules:       map[string]map[string][]byte{},
		},
		SecurityConfig: SecurityConfig{
			DisableProtection: false,
			AutoActivate:      false,
			NoLevelLimits:     false,
			BannedIPs:         []string{},
		},
	}
}
