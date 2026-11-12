package source

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// OpenPatchFile opens a patch from an HTTP(S) URL or a local path.
// A URL is cached at cacheDir/namePrefix_patch<ext>. A local path, including a file:// URL, is read in place and is not downloaded.
func OpenPatchFile(spec, cacheDir, namePrefix string) (*TSUpdatableFile, error) {
	if spec == "" {
		return nil, fmt.Errorf("patch source is empty")
	}

	lower := strings.ToLower(spec)
	switch {
	case strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "http://"):
		parsed, err := url.Parse(spec)
		if err != nil {
			return nil, fmt.Errorf("parse patch source: %w", err)
		}
		return NewTSUpdatableFile(
			filepath.Join(cacheDir, namePrefix+"_patch"+filepath.Ext(parsed.Path)),
			spec,
		), nil
	case strings.HasPrefix(lower, "file://"):
		parsed, err := url.Parse(spec)
		if err != nil {
			return nil, fmt.Errorf("parse patch source: %w", err)
		}
		if parsed.Host != "" && !strings.EqualFold(parsed.Host, "localhost") {
			return nil, fmt.Errorf("file URL %q has unsupported host %q", spec, parsed.Host)
		}
		if parsed.Path == "" {
			return nil, fmt.Errorf("file URL %q has an empty path", spec)
		}
		if err := requireLocalPatch(parsed.Path); err != nil {
			return nil, err
		}
		return NewTSUpdatableFile(parsed.Path, ""), nil
	default:
		if err := requireLocalPatch(spec); err != nil {
			return nil, err
		}
		return NewTSUpdatableFile(spec, ""), nil
	}
}

func requireLocalPatch(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("patch file %q: %w", path, err)
	}
	return nil
}
