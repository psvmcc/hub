package types

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	g.Latest = g.Versions[0]
	for _, v := range g.Versions {
		if v.Time.After(g.Latest.Time) {
			g.Latest = v
		}
	}
	return nil
}
