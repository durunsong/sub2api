package ctxkey

import "context"

type noContentPersistenceKey struct{}

// WithNoContentPersistence marks ephemeral content; billing metadata is retained.
func WithNoContentPersistence(ctx context.Context) context.Context {
	return context.WithValue(ctx, noContentPersistenceKey{}, true)
}
func NoContentPersistence(ctx context.Context) bool {
	return ctx != nil && ctx.Value(noContentPersistenceKey{}) == true
}
