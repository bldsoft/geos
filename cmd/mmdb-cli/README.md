## mmdb-cli

## Command: merge

### Usage

Run directly from the repository root:

```bash
go run github.com/bldsoft/geos/cmd/mmdb-cli merge <input1.mmdb> <input2.mmdb> ... <inputN.mmdb> <output.mmdb>
```

- `<input*.mmdb>`: one or more input MMDB files to be merged.
- `<output.mmdb>`: the merged database will be written here.

### Example

```bash
go run github.com/bldsoft/geos/cmd/mmdb-cli merge \
  HostingRangesIPv4.mmdb \
  HostingRangesIPv6.mmdb \
  merged.mmdb
```

This merges `HostingRangesIPv4.mmdb` and `HostingRangesIPv6.mmdb` into `merged.mmdb`.

