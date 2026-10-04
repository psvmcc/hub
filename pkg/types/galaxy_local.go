package types

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type VersionInfo struct {
	Version  string
	Time     time.Time
	Size     int64
	Filename string
}

type GalaxyLocal struct {
	Versions []VersionInfo
	Latest   VersionInfo
}

func (g *GalaxyLocal) List(dest, namespace, name string) error {
	pattern := fmt.Sprintf(`%s-%s-(\d+\.\d+\.\d+)\.tar.gz`, namespace, name)
	re := regexp.MustCompile(pattern)
	var v []string
	err := filepath.Walk(dest, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		matches := re.FindStringSubmatch(info.Name())
		if len(matches) > 1 {
			var vi VersionInfo
			vi.Version = matches[1]
			vi.Time = info.ModTime()
			vi.Size = info.Size()
			vi.Filename = info.Name()
			g.Versions = append(g.Versions, vi)
			v = append(v, vi.Version)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to parse directory %s, got error: %s", dest, err)
	}

	if len(g.Versions) == 0 {
		return fmt.Errorf("no versions found in directory %s", dest)
	}

	g.Latest = g.Versions[0]
	for _, v := range g.Versions {
		if compareVersions(v.Version, g.Latest.Version) > 0 {
			g.Latest = v
		}
	}
	return nil
}

// compareVersions compares two MAJOR.MINOR.PATCH versions numerically.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := range min(len(as), len(bs)) {
		x, y := strings.TrimLeft(as[i], "0"), strings.TrimLeft(bs[i], "0")
		if c := cmp.Compare(len(x), len(y)); c != 0 {
			return c
		}
		if c := cmp.Compare(x, y); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(as), len(bs))
}
