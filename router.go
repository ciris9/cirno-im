package cim

import (
	"cirno-im/wire/pkt"
	"errors"
	"sync"
)

var ErrSessionLost = errors.New("err: session lost")

type Router struct {
	handlers *FuncTree
	pool     sync.Pool
}

func NewRouter() *Router {
	r := &Router{
		handlers: NewTree(),
	}
	r.pool.New = func() interface{} {
		return BuildContext()
	}
	return r
}

func (r *Router) Handle(command string, handlers ...HandlerFunc) {
	r.handlers.Add(command, handlers...)
}

func (r *Router) Serve(packet *pkt.LogicPkt, dispatcher Dispatcher, cache SessionStorage, session Session) error {
	if dispatcher == nil {
		return errors.New("dispacher is nil")
	}
	if cache == nil {
		return errors.New("cache is nil")
	}
	ctx := r.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.request = packet
	ctx.Dispatcher = dispatcher
	ctx.SessionStorage = cache
	ctx.session = session
	r.serveContext(ctx)
	r.pool.Put(ctx)
	return nil
}

func (r *Router) serveContext(ctx *ContextImpl) {
	chain, ok := r.handlers.Get(ctx.Header().Command)
	if !ok {
		ctx.handlers = []HandlerFunc{handleNoFound}
		ctx.Next()
		return
	}
	ctx.handlers = chain
	ctx.Next()
}

func handleNoFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResponse{Message: "NoImplemented"})
}

type FuncTree struct {
	nodes map[string]HandlerChain
}

func NewTree() *FuncTree {
	return &FuncTree{
		nodes: map[string]HandlerChain{},
	}
}

func (t *FuncTree) Add(path string, handlers ...HandlerFunc) {
	t.nodes[path] = append(t.nodes[path], handlers...)
}

func (t *FuncTree) Get(path string) (HandlerChain, bool) {
	f, ok := t.nodes[path]
	return f, ok
}
