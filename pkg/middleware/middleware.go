package middleware

import "go.uber.org/zap"

type MiddlewareBase struct {
	s *zap.SugaredLogger
}

type Middleware[TOutput any] interface {
	Process() (TOutput, error)
}
