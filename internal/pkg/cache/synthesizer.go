package cache

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/allegro/bigcache"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type synthesizer interface {
	Work(string) (*api.Result, error)
}

//BigCacher keeps cached results
type BigCacher struct {
	realSynt synthesizer
	cache    *bigcache.BigCache
}

//NewCacher creates cached worker
func NewCacher(rw synthesizer, config *viper.Viper) (*BigCacher, error) {
	if rw == nil {
		return nil, errors.New("No synthesizer")
	}
	res := &BigCacher{}
	res.realSynt = rw
	dur := config.GetDuration("duration")
	if dur > 0 {
		cfg := bigcache.DefaultConfig(dur)
		cfg.CleanWindow = getCleanDuration(config.GetDuration("cleanDuration"))
		cfg.Logger = goapp.Log
		var err error
		res.cache, err = bigcache.NewBigCache(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "Can't init cache")
		}
		goapp.Log.Infof("Cache initialized with duration: %s", dur.String())
	} else {
		goapp.Log.Infof("No cache initialized")
	}
	return res, nil
}

//Work try find in cache or invoke real worker
func (c *BigCacher) Work(text string) (*api.Result, error) {
	if c.cache == nil {
		return c.realSynt.Work(text)
	}

	entry, err := c.cache.Get(text)
	if err == nil {
		goapp.Log.Debug("Found in cache")
		return &api.Result{AudioAsString: string(entry)}, nil
	}
	goapp.Log.Debug("Not found in cache")
	res, err := c.realSynt.Work(text)
	if err == nil && len(res.ValidationFailures) == 0 {
		c.cache.Set(text, []byte(res.AudioAsString))
	}
	return res, err
}

func getCleanDuration(dur time.Duration) time.Duration {
	if dur > 0 {
		return dur
	}
	return 5 * time.Minute
}
