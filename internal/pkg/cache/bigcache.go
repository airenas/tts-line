package cache

import (
	"context"
	"fmt"
	"hash/fnv"
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
		goapp.Log.Info().Str("duration", dur.String()).Str("cleanDuration", cfg.CleanWindow.String()).Msg("Cache initialized")
		if cfg.HardMaxCacheSize > 0 {
			goapp.Log.Info().Int("value", cfg.HardMaxCacheSize).Msg("Cache max memory in MB")
		}
		goapp.Log.Info().Int("value", res.maxTextLen).Msg("Cache max len for caching text")
	} else {
		goapp.Log.Info().Msg("No cache initialized")
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

	k := key(inp)
	entry, err := c.cache.Get(k)
	if err == nil {
		log.Ctx(ctx).Debug().Msg("Found in cache")
		return &api.Result{Audio: entry}, nil
	}
	log.Ctx(ctx).Debug().Msg("Not found in cache")
	res, err := c.realSynt.Work(ctx, inp)
	if res != nil && err == nil {
		_ = c.cache.Set(k, res.Audio)
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
	h := fnv.New64a()
	h.Write([]byte(inp.Text))
	h.Write([]byte(inp.OutputFormat.String()))
	h.Write([]byte(inp.Voice))
	h.Write([]byte(fmt.Sprintf("%.4f", inp.Speed)))
	h.Write([]byte(strconv.FormatInt(inp.MaxEdgeSilenceMillis, 10)))
	h.Write([]byte(inp.SymbolMode))
	h.Write([]byte(fmt.Sprintf("%s", inp.SelectedSymbols)))
	//	return inp.Text + "_" + inp.OutputFormat.String() + "_" + fmt.Sprintf("%.4f", inp.Speed) + "_" + inp.Voice + "_" + strconv.FormatInt(inp.MaxEdgeSilenceMillis, 10)
	return strconv.FormatUint(h.Sum64(), 36)
}
