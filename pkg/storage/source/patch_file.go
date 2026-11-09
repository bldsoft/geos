package source

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// OpenPatchFile opens a patch from an HTTP(S) URL or a local path.
// A URL is cached at cacheDir/namePrefix_patch<ext>. A local path, including a file:// URL, is read in place and is not downloaded.
func OpenPatchFile(spec, cacheDir, namePrefix string) (*TSUpdatableFile, error) {
	if spec == "" {
		return nil, fmt.Errorf("patch source is empty")
	}

	parsed, err := url.Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("parse patch source: %w", err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return NewTSUpdatableFile(
			filepath.Join(cacheDir, namePrefix+"_patch"+filepath.Ext(parsed.Path)),
			spec,
		), nil
	case "file":
		if parsed.Path == "" {
			return nil, fmt.Errorf("file URL %q has an empty path", spec)
		}
		return NewTSUpdatableFile(parsed.Path, ""), nil
	case "":
		return NewTSUpdatableFile(spec, ""), nil
	default:
		return nil, fmt.Errorf("unsupported patch source %q", spec)
	}
}
