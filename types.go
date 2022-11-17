package tiny

import "time"

// ToPtr converts any value to a pointer
func ToPtr[V any](v V) *V {
	return &v
}

// TimePtrToUnix returns a pointer to epoch seconds in UTC timezone represented by given pointer to time.Time
func TimePtrToUnix(t *time.Time) *int64 {
	if t == nil {
		return nil
	}

	value := (*t).UTC().Unix()
	return &value
}

// UnixPtrToTime returns a pointer to time.Time by converting pointer to epoch seconds in UTC timezone
func UnixPtrToTime(u *int64) *time.Time {
	if u == nil {
		return nil
	}

	value := time.Unix(*u, 0).UTC()
	return &value
}

// TimeToUnixPtr returns a pointer to epoch seconds in UTC timezone represented by given time.Time
func TimeToUnixPtr(t time.Time) *int64 {
	value := t.UTC().Unix()
	return &value
}

// UnixToTimePtr returns a pointer to time.Time by converting epoch seconds in UTC timezone
func UnixToTimePtr(u int64) *time.Time {
	value := time.Unix(u, 0).UTC()
	return &value
}
