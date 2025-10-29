package cache

import (
	"context"
	"fmt"
	slog "log"
	"strconv"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/service"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/allegro/bigcache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// BigCacher keeps cached results
type BigCacher struct {
	realSynt   service.Synthesizer
	cache      *bigcache.BigCache
	maxTextLen int // do not add to cache bigger results
}

// NewCacher creates cached worker
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
		cfg.Logger = slog.New(goapp.Log, "", 0)

		cfg.Shards = 64
		cfg.HardMaxCacheSize = config.GetInt("maxMB")
		if cfg.HardMaxCacheSize > 0 {
			cfg.MaxEntriesInWindow = cfg.HardMaxCacheSize * 1024
		}
		var err error
		res.cache, err = bigcache.NewBigCache(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "Can't init cache")
		}
		res.maxTextLen = config.GetInt("maxTextLen")
		goapp.Log.Info().Msgf("Cache initialized with duration: %s, clean duration: %s", dur.String(), cfg.CleanWindow.String())
		if cfg.HardMaxCacheSize > 0 {
			goapp.Log.Info().Msgf("Cache max memomy in MB %d", cfg.HardMaxCacheSize)
		}
		goapp.Log.Info().Msgf("Cache max len for caching text %d", res.maxTextLen)
	} else {
		goapp.Log.Info().Msgf("No cache initialized")
	}
	return res, nil
}

// Work try find data in the cache or invoke a real worker
func (c *BigCacher) Work(ctx context.Context, inp *api.TTSRequestConfig) (*api.Result, error) {
	ctx, span := utils.StartSpan(ctx, "BigCacher.Work")
	defer span.End()

	if c.cache == nil || !c.isOK(inp) {
		return c.realSynt.Work(ctx, inp)
	}

	entry, err := c.cache.Get(key(inp))
	if err == nil {
		log.Ctx(ctx).Debug().Msg("Found in cache")
		return &api.Result{Audio: entry}, nil
	}
	log.Ctx(ctx).Debug().Msg("Not found in cache")
	res, err := c.realSynt.Work(ctx, inp)
	if res != nil && err == nil {
		_ = c.cache.Set(key(inp), res.Audio)
	}
	return res, err
}

func (c *BigCacher) isOK(inp *api.TTSRequestConfig) bool {
	return (c.maxTextLen == 0 || len(inp.Text) <= c.maxTextLen) && inp.OutputTextFormat == api.TextNone && len(inp.SpeechMarkTypes) == 0
}

func getCleanDuration(dur time.Duration) time.Duration {
	if dur > 0 {
		return dur
	}
	return 5 * time.Minute
}

func key(inp *api.TTSRequestConfig) string {
	return inp.Text + "_" + inp.OutputFormat.String() + "_" + fmt.Sprintf("%.4f", inp.Speed) + "_" + inp.Voice + "_" + strconv.FormatInt(inp.MaxEdgeSilenceMillis, 10)
}
