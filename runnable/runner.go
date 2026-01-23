package runnable

import (
	"context"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/x893675/valhalla-common/logger"
)

type RunnableService interface {
	Run(ctx context.Context) error
}

type RunnableFunc func(ctx context.Context) error

func (f RunnableFunc) Run(ctx context.Context) error {
	return f(ctx)
}

type NamedRunnableService interface {
	RunnableService

	Name() string
}

type Runner interface {
	RunServices(ctx context.Context, services ...RunnableService) error
}

type RunnerFunc func(ctx context.Context, services ...RunnableService) error

func (f RunnerFunc) RunServices(ctx context.Context, services ...RunnableService) error {
	return f(ctx, services...)
}

func RunServices(ctx context.Context, services ...RunnableService) error {
	return NewRunner().RunServices(ctx, services...)
}

type RunnerOption func(r *runner)

type ErrorHandler func(service RunnableService, err error) error

type runner struct {
	logger        logger.Logger
	errorHandler  ErrorHandler
	errorInterval time.Duration
}

func NewRunner(options ...RunnerOption) Runner {
	r := &runner{
		logger: logger.WithName("runnable"),
		errorHandler: func(service RunnableService, err error) error {
			return err
		},
		errorInterval: 20 * time.Second,
	}

	for _, option := range options {
		option(r)
	}

	return r
}

func (r *runner) RunServices(ctx context.Context, services ...RunnableService) error {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		// safe cancel
		if ctx.Err() == nil {
			cancel()
		}
	}()

	errChan := make(chan error)
	defer close(errChan)

	for _, service := range services {
		go func(ctx context.Context, service RunnableService) {
			for {
				select {
				case <-ctx.Done():
					return

				default:
					if err := service.Run(ctx); err != nil {
						if err = r.errorHandler(service, err); err != nil {
							if ctx.Err() == nil {
								// safe push
								select {
								case errChan <- err:
								default:
								}
							}
							return
						}
					}
					time.Sleep(r.errorInterval)
				}
			}
		}(ctx, service)

	}

	select {
	case <-ctx.Done():

	case err := <-errChan:
		// only return the first error
		return err
	}
	return nil
}

func getServiceName(s RunnableService) string {
	if ns, ok := s.(NamedRunnableService); ok {
		return ns.Name()
	}
	return reflect.TypeOf(s).String()
}

func WithLogger(logger logger.Logger) RunnerOption {
	return func(r *runner) {
		r.logger = logger
	}
}

func WithErrorHandler(errorHandler ErrorHandler) RunnerOption {
	return func(r *runner) {
		r.errorHandler = errorHandler
	}
}

func LogOnError() RunnerOption {
	return func(r *runner) {
		r.errorHandler = func(service RunnableService, err error) error {
			r.logger.WithFields(
				zap.String("svc", getServiceName(service)),
				zap.Error(err),
			).Errorf("Service failed, retry after %v", r.errorInterval)
			return nil
		}
	}
}

func WithErrorInterval(interval time.Duration) RunnerOption {
	return func(r *runner) {
		r.errorInterval = interval
	}
}
