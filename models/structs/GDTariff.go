package structs

type GDTariff struct {
	Title    string
	PriceRUB int
	PriceUSD float64

	Players  int
	Levels   int
	Comments int
	Posts    int

	CustomMusic bool
	Music       GDTariffMusic
	GDLab       GDTariffBuildlab
	ACL         bool
	Shops       bool
	Roles       bool

	Modules    bool
	Backups    bool
	Logs       bool
	Levelpacks bool
	Quests     bool
}

type GDTariffMusic struct {
	YouTube bool
	Deezer  bool
	VK      bool
	Files   bool
}

type GDTariffBuildlab struct {
	Enabled  bool
	IOS      bool
	MacOS    bool
	Icons    bool
	Textures bool
	MultiVer bool
	Extended bool
}
