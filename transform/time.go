package transform

import (
	"fmt"
	"time"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

const (
	timeDefaultFmt = "2006-01-02T15:04:05.000Z"
)

type timeUnixConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *timeUnixConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *timeUnixConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

type timePatternConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`

	Format   string `json:"format"`
	Location string `json:"location"`
}

func (c *timePatternConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *timePatternConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Format == "" {
		return fmt.Errorf("format: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func timeUnixToBytes(t time.Time) []byte {
	return []byte(fmt.Sprintf("%d", t.UnixNano()))
}

// timeUnixToStr converts a UnixNano timestamp to a string.
func timeUnixToStr(ts int64, timeFmt string, loc string) (string, error) {
	timeDate := time.Unix(0, ts)

	if loc != "" {
		ll, err := time.LoadLocation(loc)
		if err != nil {
			return "", fmt.Errorf("location %s: %v", loc, err)
		}

		timeDate = timeDate.In(ll)
	}

	return timeDate.Format(timeFmt), nil
}

func timeStrToUnix(timeStr, timeFmt string, loc string) (time.Time, error) {
	var timeDate time.Time
	if loc != "" {
		ll, err := time.LoadLocation(loc)
		if err != nil {
			return timeDate, fmt.Errorf("location %s: %v", loc, err)
		}

		pil, err := time.ParseInLocation(timeFmt, timeStr, ll)
		if err != nil {
			return timeDate, fmt.Errorf("format %s location %s: %v", timeFmt, loc, err)
		}

		timeDate = pil
	} else {
		p, err := time.Parse(timeFmt, timeStr)
		if err != nil {
			return timeDate, fmt.Errorf("format %s: %v", timeFmt, err)
		}

		timeDate = p
	}

	return timeDate, nil
}
