package structs

type MCTariff struct {
	Title    string
	PriceRUB int

	CPU       uint
	MinRamGB  uint
	MaxRamGB  uint
	IsDynamic bool
	DiskGB    uint
}

func (m MCTariff) GetStartupTemplate() string {
	if m.IsDynamic {
		return "java -Xms128M -Xmx{{SERVER_MEMORY}}M -Dterminal.jline=false -Dterminal.ansi=true -XX:+UseShenandoahGC -XX:+UnlockExperimentalVMOptions -XX:ShenandoahUncommitDelay=10000 -XX:ShenandoahGuaranteedGCInterval=60000 -jar {{SERVER_JARFILE}}"
	} else {
		return "java -Xms128M -Xmx{{SERVER_MEMORY}}M -jar {{SERVER_JARFILE}}"
	}
}

func (m MCTariff) GetRAM() uint {
	if m.IsDynamic {
		return m.MaxRamGB
	} else {
		return m.MaxRamGB
	}
}

func (m MCTariff) GetSwap() uint {
	if m.IsDynamic {
		return 0
	} else {
		return 0
	}
}

var MCCoresEggs = map[string]MCCore{
	"vanilla": {
		Title:             "Vanilla",
		EggID:             5,
		VersionField:      "VANILLA_VERSION",
		VersionConstraint: ">= 1.2.1, < 1.21",
	},
	"spigot": {
		Title:             "Spigot",
		EggID:             60,
		VersionField:      "DL_VERSION",
		VersionConstraint: ">= 1.4.6, < 1.21",
	},
	"paper": {
		Title:             "Paper",
		EggID:             75,
		VersionField:      "MINECRAFT_VERSION",
		VersionConstraint: ">= 1.8.8, < 1.21",
	},
	"purpur": {
		Title:             "Purpur",
		EggID:             85,
		VersionField:      "MINECRAFT_VERSION",
		VersionConstraint: ">= 1.14.1, < 1.21",
	},
	"folia": {
		Title:             "Folia",
		EggID:             80,
		VersionField:      "MINECRAFT_VERSION",
		VersionConstraint: ">= 1.19.4, <= 1.20.2",
	},
	"fabric": {
		Title:             "Fabric",
		EggID:             65,
		VersionField:      "MC_VERSION",
		VersionConstraint: ">= 1.14, <= 1.21",
	},
	"forge": {
		Title:             "Forge",
		EggID:             104,
		VersionField:      "MC_VERSION",
		VersionConstraint: ">= 1.2.3, < 1.21",
	},
	"quilt": {
		Title:             "Quilt",
		EggID:             70,
		VersionField:      "MC_VERSION",
		VersionConstraint: ">= 1.14.4, < 1.21",
	},
	"sponge": {
		Title:             "Sponge",
		EggID:             110,
		VersionField:      "SV_VERSION",
		VersionConstraint: ">= 1.8, <= 1.20.2",
	},
	"spongeforge": {
		Title:             "SpongeForge",
		EggID:             116,
		VersionField:      "SF_VERSION",
		VersionConstraint: ">= 1.8, <= 1.16.5",
	},
}

var MCDockerImages = map[string]string{
	">= 1.20.5":         "Java 21",
	">= 1.17, < 1.20.5": "Java 17",
	"< 1.17":            "Java 8",
}
