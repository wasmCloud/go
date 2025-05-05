package events

import (
	"context"
	"sync"

	"go.wasmcloud.dev/x/wasmbus"
)

// Router is a multiplexer for events. It routes events to the appropriate handlers based on the event type.
// Use the `AddRoute` and `RemoveRoute` methods to add and remove routes, these are safe for concurrent use.
// Safe to be created manually or via the `NewEventRouter` function.
type Router struct {
	routes sync.Map
}

var _ EventHandler = (*Router)(nil)

// NewEventRouter creates a new EventRouter.
// It demultiplexes events to the appropriate handlers based on the event type, using a single event bus subscription.
func NewEventRouter() *Router {
	return &Router{}
}

// Route creates a new EventRouteHandler for the given event type.
func Route[T any](callback func(context.Context, T)) EventHandler {
	return DiscardErrorsHandler(func(ctx context.Context, ev Event) {
		typedEvent, ok := ev.BusEvent.(T)
		if !ok {
			return
		}

		callback(ctx, typedEvent)
	})
}

// AddRoute adds a new route to the event router.
func (r *Router) AddRoute(identifier string, route EventHandler) {
	r.routes.LoadOrStore(identifier, route)
}

// RemoveRoute removes a route from the event router.
func (r *Router) RemoveRoute(identifier string) {
	r.routes.Delete(identifier)
}

// HandleEvent implements the EventHandler interface
func (r *Router) HandleEvent(ctx context.Context, event Event) {
	r.routes.Range(func(_ any, value any) bool {
		handler := value.(EventHandler)
		handler.HandleEvent(ctx, event)
		return true
	})
}

// HandleError implements the EventHandler interface
func (r *Router) HandleError(context.Context, *wasmbus.Message, error) {
	// serialization errors are discarded
}
