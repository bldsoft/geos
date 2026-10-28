package entity

type ASN struct {
	ASN uint32 `json:"asn"`
	ASO string `json:"aso"`
}

type ASNFilter struct {
	NamePrefix string   `schema:"name-prefix" json:"namePrefix"`
	Limit      uint32   `schema:"limit" json:"limit"`
	Numbers    []uint32 `schema:"asns" json:"asns"`
}
