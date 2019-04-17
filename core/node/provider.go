package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-routing"
	"go.uber.org/fx"

	"github.com/ipfs/go-ipfs/pin"
	"github.com/ipfs/go-ipfs/provider"
	"github.com/ipfs/go-ipfs/provider/deprecated"
	q "github.com/ipfs/go-ipfs/provider/queue"
	"github.com/ipfs/go-ipfs/repo"
)

const kReprovideFrequency = time.Hour * 12

func ProviderQueue(mctx MetricsCtx, lc fx.Lifecycle, repo repo.Repo) (*q.Queue, error) {
	return q.NewQueue(lifecycleCtx(mctx, lc), "provider-v1", repo.Datastore())
}

func ProviderCtor(mctx MetricsCtx, lc fx.Lifecycle, queue *q.Queue, rt routing.IpfsRouting) provider.Provider {
	p := deprecated.NewProvider(lifecycleCtx(mctx, lc), queue, rt)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return p.Close()
		},
	})

	return p
}

func ReproviderCtor(mctx MetricsCtx, lc fx.Lifecycle, cfg *config.Config, bs BaseBlocks, ds format.DAGService, pinning pin.Pinner, rt routing.IpfsRouting) (*deprecated.Reprovider, error) {
	var keyProvider deprecated.KeyChanFunc

	switch cfg.Reprovider.Strategy {
	case "all":
		fallthrough
	case "":
		keyProvider = deprecated.NewBlockstoreProvider(bs)
	case "roots":
		keyProvider = deprecated.NewPinnedProvider(pinning, ds, true)
	case "pinned":
		keyProvider = deprecated.NewPinnedProvider(pinning, ds, false)
	default:
		return nil, fmt.Errorf("unknown reprovider strategy '%s'", cfg.Reprovider.Strategy)
	}
	return deprecated.NewReprovider(lifecycleCtx(mctx, lc), rt, keyProvider), nil
}

func ProviderSysCtor(mctx MetricsCtx, lc fx.Lifecycle, p provider.Provider, r provider.Reprovider) provider.System {
  return provider.NewSystem(p, r)
}

func Reprovider(cfg *config.Config, reprovider *deprecated.Reprovider) error {
	reproviderInterval := kReprovideFrequency
	if cfg.Reprovider.Interval != "" {
		dur, err := time.ParseDuration(cfg.Reprovider.Interval)
		if err != nil {
			return err
		}

		reproviderInterval = dur
	}

	go reprovider.Run(reproviderInterval)
	return nil
}
