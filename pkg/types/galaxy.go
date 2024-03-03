package types

type APIVersions struct {
	AvailableVersions struct {
		V3 string `json:"v3"`
	} `json:"available_versions"`
}
