package entity

import "github.com/maxmind/mmdbwriter/mmdbtype"

type Hosting struct {
	Datacenter string `maxminddb:"datacenter" json:"datacenter,omitempty"`
	Domain     string `maxminddb:"domain" json:"domain,omitempty"`
}

func (h Hosting) ToMMDBType() mmdbtype.Map {
	res := make(mmdbtype.Map)
	res[mmdbtype.String("datacenter")] = mmdbtype.String(h.Datacenter)
	res[mmdbtype.String("domain")] = mmdbtype.String(h.Domain)
	return res
}

func (h Hosting) MarshalCSV() (names, row []string, err error) {
	names = []string{
		"datacenter",
		"domain",
	}
	row = []string{
		h.Datacenter,
		h.Domain,
	}
	return names, row, nil
}
