package convert

import "time"

func StringToTime(s string) (time.Time, error) {
	layout := time.RFC3339
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
func PtrTime(t time.Time) *time.Time {
	return &t
}
