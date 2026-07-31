//go:build !linux

package sandbox

import "strconv"

func Apply(port string) error {
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return err
	}
	return nil
}
