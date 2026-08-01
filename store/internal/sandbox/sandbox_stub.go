//go:build !linux

package sandbox

func Apply(port, dataDir string) error {
	return nil
}
