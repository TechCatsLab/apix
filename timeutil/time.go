/*
 * Revision History:
 *     Initial: 2018/08/18        Feng Yifei
 */

package timeutil

import (
	"time"
)

// TimeToMillis converts a time.Time to millisecond.
func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
