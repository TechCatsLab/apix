// Reference: https://github.com/mholt/caddy

package runtime

import (
	"errors"
	rt "runtime"
	"strconv"
	"strings"
)

// SetCPU parses string cpu and sets GOMAXPROCS
// according to its value. It accepts either
// a number (e.g. 3) or a percent (e.g. 50%).
// If the percent resolves to less than a single
// GOMAXPROCS, it rounds it up to GOMAXPROCS=1.
func SetCPU(cpu string) error {
	var numCPU int

	availCPU := rt.NumCPU()

	if strings.HasSuffix(cpu, "%") {
		// Percent
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
		if numCPU < 1 {
			numCPU = 1
		}
	} else {
		// Number
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	rt.GOMAXPROCS(numCPU)
	return nil
}
