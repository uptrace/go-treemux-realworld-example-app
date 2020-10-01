package rwe

import (
	"context"
	mathRand "math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/uptrace/go-realworld-example-app/xconfig"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

var Clock = clock.New()

var (
	WaitGroup sync.WaitGroup
	ExitCh    = make(chan struct{})
	exiting   uint32
)

func Exiting() bool {
	return atomic.LoadUint32(&exiting) == 1
}

func Running() bool {
	return !Exiting()
}

//------------------------------------------------------------------------------

var (
	Config *xconfig.Config
	Ctx    context.Context
)

func Init(ctx context.Context, cfg *xconfig.Config) context.Context {
	if Config != nil {
		panic("not reached")
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	mathRand.Seed(time.Now().UnixNano())

	Config = cfg
	Ctx = ctx

	callOnInit(ctx)
	setupOtel(ctx)

	return ctx
}

//------------------------------------------------------------------------------

type hookFn func(context.Context)

var onInit []hookFn

func OnInit(fn hookFn) {
	if Ctx != nil {
		fn(Ctx)
		return
	}
	onInit = append(onInit, fn)
}

func callOnInit(ctx context.Context) {
	run(ctx, onInit)
	onInit = nil
}

//------------------------------------------------------------------------------

var (
	primarily []hookFn
	secondary []hookFn
)

func Exit(ctx context.Context) {
	if !atomic.CompareAndSwapUint32(&exiting, 0, 1) {
		return
	}

	close(ExitCh)
	if waitTimeout(&WaitGroup, 30*time.Second) {
		logrus.WithContext(ctx).Info("waitTimeout")
	}

	run(ctx, primarily)
	run(ctx, secondary)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func run(ctx context.Context, hookFns []hookFn) {
	var wg sync.WaitGroup
	wg.Add(len(hookFns))
	for _, h := range hookFns {
		go func(h hookFn) {
			defer wg.Done()
			h(ctx)
		}(h)
	}
	wg.Wait()
}

func OnExit(h hookFn) {
	primarily = append(primarily, h)
}

func OnExitSecondary(h hookFn) {
	secondary = append(secondary, h)
}

//------------------------------------------------------------------------------

func WaitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}

func IsDebug() bool {
	switch Config.Env {
	case "prod":
		return false
	}
	return true
}
