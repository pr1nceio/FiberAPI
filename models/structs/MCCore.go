package structs

import "github.com/hashicorp/go-version"

type MCCore struct {
	Title             string `json:"title"`
	EggID             int    `json:"egg_id"`
	VersionField      string `json:"version"`
	VersionConstraint string `json:"version_constraint"`
}

func (m MCCore) VersionConstraintAsSemVer() (constraint version.Constraints, err error) {
	return version.NewConstraint(m.VersionConstraint)
}
