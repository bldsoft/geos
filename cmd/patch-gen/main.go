package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"

	"github.com/manifoldco/promptui"
)

const FirstCustomGeonameID = 1_000_000_000

var ErrGeonameIDAlreadyInUse = fmt.Errorf("geoname already in use")

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
				Name:  "city",
				Usage: "add record to custom city db file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "f",
						Value: "city_custom.json",
					},
					&cli.StringFlag{
						Name:    "geonames",
						Value:   "geonames_custom.json",
						Aliases: []string{"geonames-patch"},
					},
				},
				Action: func(ctx *cli.Context) error {
					filename := ctx.String("f")
					currentDB := make(map[string]*entity.City)
					if err := readCurrentDB(ctx.Context, filename, &currentDB); err != nil {
						return err
					}

					geonameStorage := geonamesStorage(ctx.Context, ctx.String("geonames"))

					network, err := promptNetwork(ctx.Context)
					if err != nil {
						return err
					}

					var dbCity *entity.City

					dbCity, err = readCityInput(ctx, geonameStorage)
					if err != nil {
						return err
					}

					// log
					maxmindRecord := map[string]*entity.City{
						network: dbCity,
					}
					data, err := json.MarshalIndent(maxmindRecord, "", "    ")
					if err != nil {
						return err
					}
					println(string(data))
					//

					if !areYouSure(filename) {
						return nil
					}

					currentDB[network] = dbCity
					return writeFile(ctx.Context, filename, currentDB)
				},
			},
			{
				Name:  "geonames",
				Usage: "add custom geonames to a patch file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "f",
						Value: "geonames_custom.json",
					},
					&cli.UintFlag{
						Name:  "id",
						Value: 0,
					},
				},
				Action: func(ctx *cli.Context) (err error) {
					filename := ctx.String("f")
					var records []geonames.CustomGeonamesRecord
					if err := readCurrentDB(ctx.Context, filename, &records); err != nil {
						return err
					}

					geonameStorage := geonamesStorage(ctx.Context, filename)

					geoname := ctx.Args().First()
					geonameID := ctx.Uint64("id")

					if len(geoname) > 0 && geonameID == 0 {
						geonameID = generateGeonameID(records)
					}

					if len(geoname) == 0 {
						geoname, err = geonameInput()
						if err != nil {
							return err
						}
					}

					if geonameID == 0 {
						geonameID, err = geonameIDInput(ctx.Context, geonameStorage, records)
						if err != nil {
							return err
						}
					}

					// log
					rec := geonames.NewCustomGeonamesRecord(geoname, int(geonameID))
					data, err := json.MarshalIndent(rec, "", "    ")
					if err != nil {
						return err
					}
					println(string(data))
					//

					if !areYouSure(filename) {
						return nil
					}

					records = append(records, rec)
					return writeFile(ctx.Context, filename, records)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func geonamesStorage(ctx context.Context, customFilePath string) geonames.Storage {
	originalStorage := geonames.NewStorage("/tmp/")
	geonameStorage := geonames.NewMultiStorage(originalStorage)

	if customStorage, err := geonames.NewCustomStorageFromFile(customFilePath); err == nil {
		return geonameStorage.Add(customStorage)
	}
	originalStorage.WaitReady()
	return geonameStorage
}

func isGeoNameIDAlradyInUse(ctx context.Context, storage geonames.Storage, id uint64) error {
	for _, contient := range storage.Continents(ctx) {
		if contient.GetGeoNameID() == int(id) {
			return ErrGeonameIDAlreadyInUse
		}
	}

	subdivisions, err := storage.Subdivisions(ctx, entity.GeoNameFilter{})
	if err != nil {
		return err
	}
	for _, sd := range subdivisions {
		if sd.GetGeoNameID() == int(id) {
			return ErrGeonameIDAlreadyInUse
		}
	}

	countries, err := storage.Countries(ctx, entity.GeoNameFilter{})
	if err != nil {
		return err
	}
	for _, country := range countries {
		if country.GetGeoNameID() == int(id) {
			return ErrGeonameIDAlreadyInUse
		}
	}
	cities, err := storage.Cities(ctx, entity.GeoNameFilter{})
	if err != nil {
		return err
	}
	for _, city := range cities {
		if city.GetGeoNameID() == int(id) {
			return ErrGeonameIDAlreadyInUse
		}
	}
	return nil
}

func generateGeonameID(records []geonames.CustomGeonamesRecord) uint64 {
	geonameID := uint64(FirstCustomGeonameID)
	for _, rec := range records {
		geonameID = max(geonameID, uint64(rec.City.GeoNameID)+1)
	}
	return geonameID
}

func geonameIDInput(ctx context.Context, storage geonames.Storage, records []geonames.CustomGeonamesRecord) (uint64, error) {
	prompt := promptui.Prompt{
		Label: "Enter geoname id (leave empty for auto generation)",
		Validate: func(idStr string) error {
			if len(idStr) == 0 {
				return nil
			}
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return err
			}

			return isGeoNameIDAlradyInUse(ctx, storage, id)
		},
	}
	idStr, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	if idStr == "" {
		return generateGeonameID(records), nil
	}
	return strconv.ParseUint(idStr, 10, 64)
}

func geonameInput() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter geoname",
	}
	return prompt.Run()
}

func areYouSure(filename string) bool {
	yes := []string{"yes", "y"}
	no := []string{"no", "n"}
	surePrompt := promptui.Prompt{
		Label: "Sure want to add record to " + filename + "? (yes/no)",
		Validate: func(s string) error {
			if slices.Contains(append(yes, no...), s) {
				return nil
			}
			return fmt.Errorf("type 'yes' or 'no'")
		},
		HideEntered: true,
	}
	res, err := surePrompt.Run()
	if err != nil {
		return false
	}
	return slices.Contains(yes, res)
}

func readCityInput(ctx *cli.Context, storage geonames.Storage) (*entity.City, error) {
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

func readCurrentDB[T any](ctx context.Context, filename string, out T) (err error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, out)
}

func writeFile[T any](ctx context.Context, filename string, data T) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	dataBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(dataBytes)
	return err
}

func promptNetwork(ctx context.Context) (string, error) {
	networkPrompt := promptui.Prompt{
		Label: "Enter network (IP/CIDR)",
		Validate: func(s string) error {
			_, _, err := net.ParseCIDR(s)
			return err
		},
	}
	return networkPrompt.Run()
}

func selectCountry(ctx context.Context, storage geonames.Storage) (*entity.GeoNameCountry, error) {
	countries, err := storage.Countries(ctx, entity.GeoNameFilter{})
	if err != nil {
		return nil, err
	}
	countrySelect := promptui.Select{
		Label:             "Choose country",
		Items:             countries,
		Templates:         selectTemplates,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.HasPrefix(strings.ToLower(countries[index].GetName()), strings.ToLower(input))
		},
	}
	index, _, err := countrySelect.Run()
	if err != nil {
		return nil, err
	}
	return countries[index], nil
}

func getContinent(ctx context.Context, storage geonames.Storage, country *entity.GeoNameCountry) *entity.GeoNameContinent {
	for _, continent := range storage.Continents(ctx) {
		if continent.Code() == country.Continent {
			return continent
		}
	}
	return nil
}

func selectCity(ctx context.Context, storage geonames.Storage, country *entity.GeoNameCountry) (*entity.GeoName, error) {
	cities, err := storage.Cities(ctx, entity.GeoNameFilter{
		CountryCodes: []string{country.GetCountryCode()},
	})
	if err != nil {
		return nil, err
	}
	var filteredCities []*entity.GeoName
	for _, city := range cities {
		if city.GetCountryCode() == country.GetCountryCode() {
			filteredCities = append(filteredCities, city)
		}
	}

	if len(filteredCities) == 1 {
		return filteredCities[0], nil
	}

	citySelect := promptui.Select{
		Label:             "Choose city",
		Items:             filteredCities,
		Templates:         selectTemplates,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.HasPrefix(strings.ToLower(filteredCities[index].GetName()), strings.ToLower(input))
		},
	}
	index, _, err := citySelect.Run()
	if err != nil {
		return nil, err
	}
	return filteredCities[index], nil
}

func getSubdivisions(ctx context.Context, storage geonames.Storage, city *entity.GeoName) ([]entity.GeoNameAdminSubdivision, error) {
	subdivisions, err := storage.Subdivisions(ctx, entity.GeoNameFilter{CountryCodes: []string{city.GetCountryCode()}})
	if err != nil {
		return nil, err
	}

	subdivisionByCode := func(adminCode string) *entity.GeoNameAdminSubdivision {
		i := slices.IndexFunc(subdivisions, func(sd *entity.GeoNameAdminSubdivision) bool {
			return adminCode == sd.AdminCode()
		})
		if i < 0 {
			return nil
		}
		return subdivisions[i]
	}

	res := make([]entity.GeoNameAdminSubdivision, 0, 4)
	for _, code := range []string{city.Admin1Code, city.Admin2Code, city.Admin3Code, city.Admin4Code} {
		if len(code) == 0 {
			break
		}
		if sd := subdivisionByCode(code); sd != nil {
			res = append(res, *sd)
		}
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
	res.Continent.GeoNameID = uint(continent.GetGeoNameID())
	res.Continent.Names = map[string]string{"en": continent.GetName()}
	res.Country.GeoNameID = uint(country.GetGeoNameID())
	// res.Country.IsInEuropeanUnion
	res.Country.IsoCode = country.GetCountryCode()
	res.Country.Names = map[string]string{"en": country.GetName()}
	res.City.GeoNameID = uint(city.GetGeoNameID())
	res.City.Names = map[string]string{"en": city.GetName()}
	//RepresentedCountry
	appendSubdvision := func(subdivisions ...entity.GeoNameAdminSubdivision) {
		for _, sd := range subdivisions {
			res.Subdivisions = append(res.Subdivisions, struct {
				GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
				IsoCode   string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
				Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
			}{
				GeoNameID: uint(sd.GetGeoNameID()),
				IsoCode:   sd.AdminCode(),
				Names:     map[string]string{"en": sd.GetName()},
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
