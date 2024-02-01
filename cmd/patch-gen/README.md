# Patch-gen
Create/Update GEOS patch for MaxMind database:
`
go run github.com/bldsoft/geos/cmd/patch-gen add
`

The patch must be placed in the directory with the original database so that Geos merges them at startup.
Patches for the city database must have a "city" prefix, for the ISP - "isp". Example: "city_custom.json"