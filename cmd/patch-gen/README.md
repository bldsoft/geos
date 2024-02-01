# Patch-gen
Create/Update GEOS patch for MaxMind database:
`
go run github.com/bldsoft/geos/cmd/patch-gen add
`
The tool will ask you to enter:
- Net. Example input: "127.0.0.1/32"
- Country. Example input: "Belarus"
- City. Example input: "Minsk"

Then it will display a preview of the entry being added and ask for confirmation to write to the file.

To add net to private network, you need to add --private flag

The patch must be placed in the directory with the original database so that Geos merges them at startup.
Patches for the city database must have a "city" prefix, for the ISP - "isp". Example: "city_custom.json"