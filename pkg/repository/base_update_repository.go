package repository

import (
	"context"
	"time"

	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"go.uber.org/atomic"
)

type updaterWithLastErr struct {
	updateFunc func(ctx context.Context, force bool) error
	inProgress atomic.Bool
	lastErr    atomic.Pointer[string]
}

func newUpdaterWithLastErr(updateFunc func(ctx context.Context, force bool) error) *updaterWithLastErr {
	return &updaterWithLastErr{
		updateFunc: updateFunc,
	}
}

func (u *updaterWithLastErr) LastErr() *string {
	return u.lastErr.Load()
}

func (u *updaterWithLastErr) InProgress() bool {
	return u.inProgress.Load()
}

func (u *updaterWithLastErr) Update(ctx context.Context, force bool) error {
	u.inProgress.Store(true)
	defer u.inProgress.Store(false)
	err := u.updateFunc(ctx, force)
	if err != nil {
		errStr := err.Error()
		u.lastErr.Store(&errStr)
		return err
	}
	u.lastErr.Store(nil)
	return nil
}

type updateOptions struct {
	force bool
	async bool
}

type baseUpdateRepository struct {
	source.LocalFileRepository
	lockFileName     string
	autoUpdatePeriod time.Duration
	updateFunc       func(ctx context.Context, force bool) error
}

func NewBaseUpdateRepository(
	lockFileName string,
	autoUpdatePeriod time.Duration,
	update func(ctx context.Context, force bool) error,
) *baseUpdateRepository {
	return &baseUpdateRepository{
		lockFileName:     lockFileName,
		autoUpdatePeriod: autoUpdatePeriod,
		updateFunc:       update,
	}
}

func (r *baseUpdateRepository) StartUpdate(ctx context.Context) error {
	return r.update(ctx, updateOptions{force: false, async: true})
}

func (r *baseUpdateRepository) update(ctx context.Context, opts updateOptions) error {
	close := func() {
		_ = r.LocalFileRepository.Remove(ctx, r.lockFileName)
	}
	if !opts.force {
		ok, unlock, err := r.TryLock(ctx, r.lockFileName)
		if !ok || err != nil {
			return err
		}
		close = unlock
	}

	if !opts.async {
		defer close()
		return r.updateFunc(ctx, opts.force)
	}

	var eg errgroup.Group
	eg.Go(func() error {
		defer close()
		ctx = context.WithoutCancel(ctx)
		return r.updateFunc(ctx, opts.force)
	})
	return nil
}

func (r *baseUpdateRepository) isInterrupted(ctx context.Context) bool {
	ok, err := r.LocalFileRepository.Exists(ctx, r.lockFileName)
	if err != nil {
		return true
	}
	return ok
}

func (r *baseUpdateRepository) Run(ctx context.Context) error {
	if r.isInterrupted(ctx) {
		ticker := time.NewTicker(time.Minute)
		for {
			err := r.update(ctx, updateOptions{force: true, async: false})
			if err == nil {
				break
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	if r.autoUpdatePeriod == 0 {
		return nil
	}

	ticker := time.NewTicker(r.autoUpdatePeriod)
	for {
		select {
		case <-ticker.C:
			err := r.update(ctx, updateOptions{force: false, async: false})
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"err": err,
				}, "failed to update")
				continue
			}
			log.FromContext(ctx).Infof("auto updated successfully")
		case <-ctx.Done():
			return nil
		}
	}
}
