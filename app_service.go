package app

import (
    rpc "github.com/philipyao/prpc/server"
    "github.com/philipyao/prpc/registry"
    "github.com/philipyao/phttp"
    "log"
)

type Service interface {
    OnInit() error
    Serve() error
    Close()
    OnFini()
}

type serviceRpc struct {
    rcvr interface{}
    name string
    addr, regAddr string
    rpcServer  *rpc.Server
}
func (sr *serviceRpc) OnInit() error {
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

func newServiceRpc(cluster string, index int, rcvr interface{},
    name, addr, registry string) *serviceRpc {
    sr := new(serviceRpc)
    sr.name = name
    sr.rcvr = rcvr
    sr.addr = addr
    sr.regAddr = registry

    rpcServer := rpc.New(
        cluster,
        index,
    )
    if rpcServer == nil {
        return nil
    }
    sr.rpcServer = rpcServer

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
