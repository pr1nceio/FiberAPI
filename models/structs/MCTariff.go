package structs

type MCTariff struct {
	Title    string
	PriceRUB uint

	CPU       uint
	MinRamGB  uint
	MaxRamGB  uint
	IsDynamic bool
	DiskGB    uint
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

var MCVersions = []string{
	"1.9.4",
	"1.12.2",
	"1.14.4",
	"1.16.5",
	"1.17.1",
	"1.18.2",
	"1.19.3",
	"1.19.4",
	"1.20.2",
}
