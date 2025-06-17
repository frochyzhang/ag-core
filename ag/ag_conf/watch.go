package ag_conf

import (
	"context"
	"sync"
)

// TODO !!!!!!!!!!!!
var onece sync.Once

type Watcher struct {
	bind        IBinder
	refreshChan chan IPropertySource
	stops       []func() error
}

func NewConfigWatcher(bind IBinder) *Watcher {

	watcher := &Watcher{
		bind:  bind,
		stops: make([]func() error, 0),
	}

	return watcher
}

func (w *Watcher) Start(context.Context) error {
	onece.Do(func() {
		env := w.bind.GetEnv()
		pslen := len(env.GetPropertySources().GetPropertySources())
		w.refreshChan = make(chan IPropertySource, pslen)
	})
	return nil
}

func (w *Watcher) Stop(context.Context) error {
	for _, stop := range w.stops {
		err := stop()
		if err != nil {
			return err
		}
	}
	return nil
}
