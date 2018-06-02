package app

import (
    "fmt"
    "os"
    "os/signal"
    "flag"
    "errors"
    "sync"
    "path/filepath"
    "syscall"

    "github.com/philipyao/toolbox/util"
    "github.com/philipyao/phttp"
)

type App struct {
    pName       string

    bInited     bool

    done        chan struct{}
    wg          sync.WaitGroup

    argIndex    *int
    argCluster  *int
    argIP       *string
    argPort     *int

    initFunc        func(chan struct{}) error
    shutdownFunc    func()
    logFunc         func(format string, args ...interface{})

    //rpc         *prpc.Worker
    http        *phttp.HTTPWorker
}

type HTTPWorker struct {
    *phttp.HTTPWorker
}

var (
    ptrWanIP          *string
)

var defaultApp  = &App{done: make(chan struct{})}
func init() {
    defaultApp.prepare()
}

func (sv *App) addr() string {
    return fmt.Sprintf("%v:%v", *sv.argIP, *sv.argPort)
}

//func (sv *App) setRpc(r *prpc.Worker) {
//    sv.rpc = r
//}
//
func (sv *App) setHttp(h *phttp.HTTPWorker) {
    sv.http = h
}

func (sv *App) prepare() {
    sv.readArgs()
}

func (sv *App) init() error {
    sv.logFunc("App start...")

    if sv.bInited {
        panic("already inited.")
    }

    err := sv.initFunc(sv.done)
    if err != nil {
        return err
    }
    sv.bInited = true
    sv.logFunc("App init ok.")
    return nil
}

func (sv *App) run() {
    if !sv.bInited {
        panic("not inited")
    }
    //if sv.rpc != nil {
    //    sv.wg.Add(1)
    //    go sv.rpc.Serve(sv.done, &sv.wg)
    //}
    if sv.http != nil {
        sv.wg.Add(1)
        go sv.http.Serve(sv.done, &sv.wg)
    }
    sv.writePid()

    sv.wg.Add(1)
    go sv.listenInterupt()

    sv.wg.Wait()

    sv.shutdownFunc()
    sv.removePid()
}

//====================================

func (sv *App) readArgs() {
    sv.argCluster = flag.Int("c", 0, "App clusterid")
    sv.argIndex = flag.Int("i", 0, "App instance index")
    sv.argIP = flag.String("l", "0.0.0.0", "App local ip")
    sv.argPort = flag.Int("p", 0, "App rpc port")
    ptrWanIP = flag.String("w", "0.0.0.0", "App wan ip")

    flag.Parse()
    if *sv.argPort <= 0 {
        panic("no App port specified!")
    }
    if *sv.argCluster <= 0 {
        panic("no App cluster id specified")
    }
}

func (sv *App) listenInterupt() {
    defer sv.wg.Done()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
    sv.shutdown()
}

func (sv *App) shutdown() {
    sv.logFunc("graceful shutdown...")
    close(sv.done)
}

func (sv *App) writePid() {
    pName := sv.processName()
    pidFile := util.GenPidFilePath(pName)
    util.WritePidToFile(pidFile, os.Getpid())
}

func (sv *App) removePid() {
    pName := sv.processName()
    pidFile := util.GenPidFilePath(pName)
    util.DeletePidFile(pidFile)
}

func (sv *App) processName() string {
    if sv.pName != "" {
        return sv.pName
    }
    sv.pName = filepath.Base(os.Args[0])
    if *sv.argIndex > 0 {
        sv.pName = fmt.Sprintf("%v%v", sv.pName, *sv.argIndex)
    }
    return sv.pName
}


//=====================================================

//必须实现，server基础接口
func HandleBase(onInit func(chan struct{}) error, onShutdown func()) error {
    if onInit == nil {
        return errors.New("nil onInit.")
    }
    if onShutdown == nil {
        return errors.New("nil onShutdown.")
    }
    defaultApp.initFunc = onInit
    defaultApp.shutdownFunc = onShutdown
    if defaultApp.logFunc == nil {
        defaultApp.logFunc = defaultLogFunc()
    }
    return defaultApp.init()
}

// 可选，注册rpc服务
//func HandleRpc(rpcName string, rpcWorker interface{}) error {
//    rpcW := prpc.New(defaultApp.addr(), rpcName, rpcWorker)
//    if rpcW == nil {
//        return errors.New(prpc.ErrMsg())
//    }
//    rpcW.SetLog(defaultApp.logFunc)
//    defaultApp.setRpc(rpcW)
//
//    return nil
//}
//
// 可选，注册http服务
func HandleHttp(addr string) (*HTTPWorker, error) {
    httpW := phttp.New(addr)
    if httpW == nil {
        return nil, errors.New("init http error")
    }
    httpW.SetLog(defaultApp.logFunc)
    defaultApp.setHttp(httpW)

    return &HTTPWorker{httpW}, nil
}

//可选，自定义log输出
func SetLogger(l func(int, string, ...interface{})) {
    defaultApp.logFunc = customLogFunc(l)
}

// server运行入口函数
func Run() {
    defaultApp.run()
}

// 获取server进程名字
func ProcessName() string {
    return defaultApp.processName()
}
