package consul

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrNoHealthyNodes  = errors.New("no healthy instances")
	ErrLabelNotFound   = errors.New("label not found")
)

const (
	META_SCHEME = "scheme"
	META_PATH   = "path"
)

type Instance struct {
	Node    string
	Address string
	Port    int

	Tags map[string]struct{}
	Meta map[string]string
}

func (i Instance) URL() (string, error) {
	scheme, ok := i.Meta[META_SCHEME]
	if !ok {
		scheme = "http"
	}
	path, ok := i.Meta[META_PATH]
	if !ok {
		path = "/"
	}

	return (&url.URL{
		Scheme: scheme,
		Host: net.JoinHostPort(
			i.Address,
			strconv.Itoa(i.Port),
		),
		Path: path,
	}).String(), nil
}

type Discovery struct {
	client      *consul.Client
	service     string
	refreshWait time.Duration

	mutex     sync.RWMutex
	instances []*Instance

	counterMutex sync.Mutex
	counter      map[string]uint64
}

type Config struct {
	Service     string
	RefreshWait time.Duration
	SkipConsul  bool
}

func New(ctx context.Context, cfg *Config) (*Discovery, error) {
	log.Ctx(ctx).Info().Str("service", cfg.Service).Dur("every", cfg.RefreshWait).Msg("init consul discovery")
	if cfg.RefreshWait <= 0 {
		cfg.RefreshWait = 30 * time.Second
	}

	d := &Discovery{
		service:     cfg.Service,
		refreshWait: cfg.RefreshWait,
		counter:     map[string]uint64{},
	}
	if !cfg.SkipConsul {
		cc := consul.DefaultConfig()
		client, err := consul.NewClient(cc)
		if err != nil {
			return nil, fmt.Errorf("discovery: %w", err)
		}
		if cfg.Service == "" {
			return nil, fmt.Errorf("discovery: no service set")
		}
		log.Ctx(ctx).Info().Str("url", cc.Address).Msg("init consul")
		d.client = client
	} else {
		log.Ctx(ctx).Warn().Msg("skip consul")
	}

	return d, nil
}

func (d *Discovery) Run(ctx context.Context) error {
	if d.client == nil {
		return fmt.Errorf("no consul")
	}
	go d.watchService(ctx)
	return nil
}

func (d *Discovery) watchService(ctx context.Context) {
	var lastIndex uint64
	log.Ctx(ctx).Info().Msg("Start AM service discovery")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		entries, meta, err := d.client.Health().Service(
			d.service,
			"",
			true, // only healthy
			&consul.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  d.refreshWait,
			},
		)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("load consul")
			time.Sleep(5 * time.Second)
			continue
		}
		log.Ctx(ctx).Debug().Int("instances", len(entries)).Msg("got from consul")
		lastIndex = meta.LastIndex

		instances, err := buildInstances(entries)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("load consul")
			time.Sleep(5 * time.Second)
			continue
		}
		d.mutex.Lock()
		d.instances = instances
		d.mutex.Unlock()

	}
}

func buildInstances(entries []*consul.ServiceEntry) ([]*Instance, error) {
	res := make([]*Instance, 0, len(entries))

	for _, entry := range entries {
		address := entry.Service.Address

		if address == "" {
			address = entry.Node.Address
		}

		if address == "" || entry.Service.Port <= 0 {
			continue
		}

		instance := &Instance{
			Node:    entry.Node.Node,
			Address: address,
			Port:    entry.Service.Port,
			Tags:    map[string]struct{}{},
			Meta:    map[string]string{},
		}
		for _, t := range entry.Service.Tags {
			instance.Tags[t] = struct{}{}
		}
		for k, v := range entry.Service.Meta {
			instance.Meta[k] = v
		}

		res = append(res, instance)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Node != res[j].Node {
			return res[i].Node < res[j].Node
		}

		if res[i].Address != res[j].Address {
			return res[i].Address < res[j].Address
		}

		return res[i].Port < res[j].Port
	})

	return res, nil
}

func (d *Discovery) URL(model string) (string, error) {
	d.mutex.RLock()
	instances := d.instances
	d.mutex.RUnlock()

	var c uint64
	for _, ins := range instances {
		if _, ok := ins.Tags[model]; ok {
			c = c + 1
		}
	}
	if c == 0 {
		return "", ErrServiceNotFound
	}
	if c == 1 {
		for _, ins := range instances {
			if _, ok := ins.Tags[model]; ok {
				return ins.URL()
			}
		}
	}

	d.counterMutex.Lock()
	defer d.counterMutex.Unlock()
	v, ok := d.counter[model]
	if !ok {
		v = 0
	}
	at := int(v % c)
	v = v + 1
	d.counter[model] = v
	i := 0
	for _, ins := range instances {
		if _, ok := ins.Tags[model]; ok {
			if i == at {
				return ins.URL()
			}
			i = i + 1
		}
	}
	return "", ErrServiceNotFound
}
