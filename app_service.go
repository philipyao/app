package app

import (
    rpc "github.com/philipyao/prpc/server"
    "github.com/philipyao/prpc/registry"
    "github.com/philipyao/phttp"
    "fmt"
    "log"
)

type Service interface {
    OnInit() error
    Serve() error
    Close()
    OnFini()
}

type serviceRpc struct {
    app *App
    rcvr interface{}
    name string
    addr, regAddr string
    rpcServer  *rpc.Server
}
func (sr *serviceRpc) OnInit() error {
    rpcServer := rpc.New(
        sr.app.Cluster(),
        sr.app.Index(),
    )
    if rpcServer == nil {
        return fmt.Errorf("create rpc error: %v %v", sr.app.Cluster(), sr.app.Index())
    }
    sr.rpcServer = rpcServer
    return sr.rpcServer.Handle(sr.rcvr, sr.name)
}
func (sr *serviceRpc) Serve() error {
    return sr.rpcServer.Serve(
        sr.addr,
        &registry.RegConfigZooKeeper{ZKAddr: sr.regAddr},
    )
}
func (sr *serviceRpc) Close() {}
func (sr *serviceRpc) OnFini() {
    sr.rpcServer.Fini()
}

func newServiceRpc(app *App, rcvr interface{},
    name, addr, regAddr string) *serviceRpc {
    sr := new(serviceRpc)
    sr.app = app
    sr.name = name
    sr.rcvr = rcvr
    sr.addr = addr
    sr.regAddr = regAddr
    return sr
}

type serviceHttp struct {
    httpServer *phttp.HTTPWorker
    handler func(worker *phttp.HTTPWorker) error
}
func (sh *serviceHttp) OnInit() error {
    return sh.handler(sh.httpServer)
}
func (sh *serviceHttp) Serve() error {
    return sh.httpServer.Serve()
}
func (sh *serviceHttp) Close() {
    sh.httpServer.Close()
}
func (sh *serviceHttp) OnFini() {}

func newServiceHttp(addr string, handler func(worker *phttp.HTTPWorker) error) *serviceHttp {
    sh := new(serviceHttp)
    httpServer := phttp.New(addr)
    if httpServer == nil {
        return nil
    }
    httpServer.SetLog(log.Printf)
    sh.httpServer = httpServer
    sh.handler = handler
    return sh
}
