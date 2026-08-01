//go:build linux

package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/landlock-lsm/go-landlock/landlock"
)

func Apply(port, dataDir string) error {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return fmt.Errorf("sandbox port: %w", err)
	}

	rules := []landlock.Rule{
		landlock.BindTCP(uint16(p)),
		landlock.RODirs(
			"/app",
			"/etc/ssl/certs",
			"/lib",
			"/usr/lib",
		).IgnoreIfMissing(),
	}

	if dataDir != "" {
		rules = append(rules, landlock.RWDirs(dataDir).IgnoreIfMissing())
	}

	if exe, err := os.Executable(); err == nil {
		rules = append(rules, landlock.ROFiles(exe))
		if dir := filepath.Dir(exe); dir != "" && dir != "/" {
			rules = append(rules, landlock.RODirs(dir))
		}
	}

	cfg := landlock.V9.BestEffort()
	if err := cfg.Restrict(rules...); err != nil {
		return fmt.Errorf("landlock restrict: %w", err)
	}
	if err := cfg.RestrictScoped(); err != nil {
		return fmt.Errorf("landlock scoped: %w", err)
	}
	return nil
}
