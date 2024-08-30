package user

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// decodeCursor takes a base64 encoded string and returns the timestamp and userID
func decodeCursor(encodedCursor string) (ts time.Time, userID string, err error) {
	cursor, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return
	}

	splitted := strings.Split(string(cursor), "|")
	if len(splitted) != 2 {
		err = fmt.Errorf("cursor is invalid")
		return
	}

	ts, err = time.Parse(time.RFC3339Nano, splitted[0])
	if err != nil {
		err = fmt.Errorf("cursor is invalid: timestamp")
		return
	}
	userID = splitted[1]
	return
}

// encodeCursor returns a base64 encoded string based on the provided timestamp and userID
func encodeCursor(ts time.Time, userID string) string {
	key := fmt.Sprintf("%s|%s", ts.Format(time.RFC3339Nano), userID)
	return base64.StdEncoding.EncodeToString([]byte(key))
}
