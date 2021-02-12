package cache

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/service"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/allegro/bigcache"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//BigCacher keeps cached results
type BigCacher struct {
	realSynt service.Synthesizer
	cache    *bigcache.BigCache
}

//NewCacher creates cached worker
func NewCacher(rw service.Synthesizer, config *viper.Viper) (*BigCacher, error) {
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
		cfg.HardMaxCacheSize = config.GetInt("maxMB")
		var err error
		res.cache, err = bigcache.NewBigCache(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "Can't init cache")
		}
		goapp.Log.Infof("Cache initialized with duration: %s, clean duration: %s", dur.String(), cfg.CleanWindow.String())
	} else {
		goapp.Log.Infof("No cache initialized")
	}
	return res, nil
}

//Work try find in cache or invoke real worker
func (c *BigCacher) Work(inp *api.TTSRequestConfig) (*api.Result, error) {
	if c.cache == nil {
		return c.realSynt.Work(inp)
	}

	entry, err := c.cache.Get(key(inp))
	if err == nil {
		goapp.Log.Debug("Found in cache")
		return &api.Result{AudioAsString: string(entry)}, nil
	}
	goapp.Log.Debug("Not found in cache")
	res, err := c.realSynt.Work(inp)
	if res != nil && err == nil && len(res.ValidationFailures) == 0 {
		c.cache.Set(key(inp), []byte(res.AudioAsString))
	}
	return res, err
}

func getCleanDuration(dur time.Duration) time.Duration {
	if dur > 0 {
		return dur
	}
	return 5 * time.Minute
}

func key(inp *api.TTSRequestConfig) string {
	return inp.Text + "_" + inp.OutputFormat.String() + "_" + inp.OutputTextFormat.String()
}
