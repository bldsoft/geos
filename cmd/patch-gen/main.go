package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"

	"github.com/manifoldco/promptui"
)

var selectTemplates = &promptui.SelectTemplates{
	Inactive: "{{ .Name }}",
	Active:   "{{ .Name | green }}",
	Selected: "{{ .Name | green }} {{ .GeoNameID | green }}",
}

func main() {
	app := &cli.App{
		Name:                 "patch-gen",
		Usage:                "GEOS DB patch generator",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add record to custom db file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "f",
						Value:   "city_custom.json",
						Aliases: []string{"file"},
					},
					&cli.BoolFlag{
						Name: "private",
					},
				},
				Action: func(ctx *cli.Context) error {
					filename := ctx.String("file")
					currentDB, err := readCurrentDB(ctx.Context, filename)
					if err != nil {
						return err
					}

					geoNameStorage := storage.NewGeoNamesStorage("/tmp/")

					network, err := promptNetwork(ctx.Context)
					if err != nil {
						return err
					}

					var dbCity *entity.City

					if ctx.Bool("private") {
						dbCity = &entity.PrivateCity
					} else {
						dbCity, err = readCityInput(ctx, geoNameStorage)
						if err != nil {
							return err
						}
					}

					geoNameStorage.WaitReady()

					// log
					maxmindRecord := map[string]*entity.City{
						network: dbCity,
					}
					data, err := json.MarshalIndent(maxmindRecord, "", "    ")
					if err != nil {
						return err
					}
					//

					currentDB[network] = dbCity

					println(string(data))
					return writeFile(ctx.Context, filename, currentDB)

				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func readCityInput(ctx *cli.Context, storage *storage.GeoNameStorage) (*entity.City, error) {
	country, err := selectCountry(ctx.Context, storage)
	if err != nil {
		return nil, err
	}

	continent := getContinent(ctx.Context, storage, country)

	city, err := selectCity(ctx.Context, storage, country)
	if err != nil {
		return nil, err
	}

	subdivisions, err := getSubdivisions(ctx.Context, storage, city)
	if err != nil {
		return nil, err
	}

	return buildDBCity(continent, country, subdivisions, city), nil
}

func readCurrentDB(ctx context.Context, filename string) (map[string]*entity.City, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	currentDB, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*entity.City)
	if len(currentDB) == 0 {
		return res, nil
	}

	if err := json.Unmarshal(currentDB, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func writeFile(ctx context.Context, filename string, db map[string]*entity.City) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(db, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func promptNetwork(ctx context.Context) (string, error) {
	networkPrompt := promptui.Prompt{
		Label: "network",
		Validate: func(s string) error {
			_, _, err := net.ParseCIDR(s)
			return err
		},
	}
	return networkPrompt.Run()
}

func selectCountry(ctx context.Context, storage *storage.GeoNameStorage) (*entity.GeoNameCountry, error) {
	countries, err := storage.Countries(ctx, entity.GeoNameFilter{})
	if err != nil {
		return nil, err
	}
	countrySelect := promptui.Select{
		Label:             "country",
		Items:             countries,
		Templates:         selectTemplates,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.HasPrefix(strings.ToLower(countries[index].Name()), strings.ToLower(input))
		},
	}
	index, _, err := countrySelect.Run()
	if err != nil {
		return nil, err
	}
	return countries[index], nil
}

func getContinent(ctx context.Context, storage *storage.GeoNameStorage, country *entity.GeoNameCountry) *entity.GeoNameContinent {
	for _, continent := range storage.Continents(ctx) {
		if continent.Code() == country.Continent {
			return continent
		}
	}
	return nil
}

func selectCity(ctx context.Context, storage *storage.GeoNameStorage, country *entity.GeoNameCountry) (*entity.GeoName, error) {
	cities, err := storage.Cities(ctx, entity.GeoNameFilter{
		CountryCodes: []string{country.CountryCode()},
	})
	if err != nil {
		return nil, err
	}
	var filteredCities []*entity.GeoName
	for _, city := range cities {
		if city.CountryCode() == country.CountryCode() {
			filteredCities = append(filteredCities, city)
		}
	}
	citySelect := promptui.Select{
		Label:             "city",
		Items:             filteredCities,
		Templates:         selectTemplates,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.HasPrefix(strings.ToLower(filteredCities[index].Name()), strings.ToLower(input))
		},
	}
	index, _, err := citySelect.Run()
	if err != nil {
		return nil, err
	}
	return filteredCities[index], nil
}

func getSubdivisions(ctx context.Context, storage *storage.GeoNameStorage, city *entity.GeoName) ([]entity.GeoNameAdminSubdivision, error) {
	subdivisions, err := storage.Subdivisions(ctx, entity.GeoNameFilter{CountryCodes: []string{city.CountryCode()}})
	if err != nil {
		return nil, err
	}

	subdivisionByCode := func(adminCode string) *entity.GeoNameAdminSubdivision {
		i := slices.IndexFunc(subdivisions, func(sd *entity.GeoNameAdminSubdivision) bool {
			return adminCode == sd.AdminCode()
		})
		return subdivisions[i]
	}

	res := make([]entity.GeoNameAdminSubdivision, 0, 4)
	for _, code := range []string{city.Admin1Code, city.Admin2Code, city.Admin3Code, city.Admin4Code} {
		if len(code) == 0 {
			break
		}
		res = append(res, *subdivisionByCode(code))
	}
	return res, nil
}

func buildDBCity(
	continent *entity.GeoNameContinent,
	country *entity.GeoNameCountry,
	subdivisions []entity.GeoNameAdminSubdivision,
	city *entity.GeoName,
) *entity.City {
	var res entity.City
	res.Continent.Code = continent.Code()
	res.Continent.GeoNameID = uint(continent.GeoNameID())
	res.Continent.Names = map[string]string{"en": continent.Name()}
	res.Country.GeoNameID = uint(country.GeoNameID())
	// res.Country.IsInEuropeanUnion
	res.Country.IsoCode = country.CountryCode()
	res.Country.Names = map[string]string{"en": country.Name()}
	res.City.GeoNameID = uint(city.GeoNameID())
	res.City.Names = map[string]string{"en": city.Name()}
	//RepresentedCountry
	appendSubdvision := func(subdivisions ...entity.GeoNameAdminSubdivision) {
		for _, sd := range subdivisions {
			res.Subdivisions = append(res.Subdivisions, struct {
				GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
				IsoCode   string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
				Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
			}{
				GeoNameID: uint(sd.GeoNameID()),
				IsoCode:   sd.AdminCode(),
				Names:     map[string]string{"en": sd.Name()},
			})
		}
	}
	appendSubdvision(subdivisions...)

	res.Location.Latitude = city.Latitude
	res.Location.Longitude = city.Longitude
	// res.Location.MetroCode
	// res.Location.AccuracyRadius
	// res.Location.TimeZone

	// res.Location.Traits

	// res.ISP *ISP `json:"ISP,omitempty"`
	return &res
}
