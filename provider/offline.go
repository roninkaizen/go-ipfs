package provider

import (
	"context"
	"github.com/ipfs/go-cid"
	"time"
)

type offlineProvider struct{}

// NewOfflineProvider creates a ProviderSystem that does nothing
func NewOfflineProvider() System {
	return &offlineProvider{}
}

func (op *offlineProvider) Run(_ time.Duration) {
}

func (op *offlineProvider) Close() error {
	return nil
}

func (op *offlineProvider) Provide(_ cid.Cid) error {
	return nil
}

func (op *offlineProvider) Reprovide(_ context.Context) error {
	return nil
}
