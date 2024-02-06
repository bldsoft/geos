# Patch-gen
Create/Update GEOS patch for MaxMind database:
`
go run github.com/bldsoft/geos/cmd/patch-gen city
`
The tool will ask you to enter:
- Net. Example input: "127.0.0.1/32"
- Country. Example input: "Belarus"
- City. Example input: "Minsk"

Then it will display a preview of the entry being added and ask for confirmation to write to the file.

To add a new geoname entity: 
`
go run github.com/bldsoft/geos/cmd/patch-gen geonames "name"
`
Generated geonames patch can be used in the city command via --geonames flag.

The patches must be placed in the directory with the original city database so that Geos merges them at startup.
Patches for the city database must have a "city" prefix, for the ISP - "isp", for the Geonames - "geonames". Example: "city_custom.json"