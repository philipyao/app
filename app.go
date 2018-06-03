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
    "log"
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

var defaultApp  = &App{done: make(chan struct{})}
func init() {
    defaultApp.prepare()
}

func (app *App) addr() string {
    return fmt.Sprintf("%v:%v", *app.argIP, *app.argPort)
}

//func (app *App) setRpc(r *prpc.Worker) {
//    app.rpc = r
//}
//
func (app *App) setHttp(h *phttp.HTTPWorker) {
    app.http = h
}

func (app *App) prepare() {
    app.prepareArgs()
}

func (app *App) init() error {
    app.logFunc("App start...")

    if app.bInited {
        panic("already inited.")
    }

    app.readArgs()

    err := app.initFunc(app.done)
    if err != nil {
        return err
    }
    app.bInited = true
    app.logFunc("App init ok.")
    return nil
}

func (app *App) run() {
    if !app.bInited {
        panic("not inited")
    }
    //if app.rpc != nil {
    //    app.wg.Add(1)
    //    go app.rpc.Serve(app.done, &app.wg)
    //}
    if app.http != nil {
        app.wg.Add(1)
        go app.http.Serve(app.done, &app.wg)
    }
    app.writePid()

    app.wg.Add(1)
    go app.listenInterupt()

    app.wg.Wait()

    app.shutdownFunc()
    app.removePid()
}

//====================================

func (app *App) prepareArgs() {
    app.argCluster = flag.Int("c", 0, "App clusterid")
    app.argIndex = flag.Int("i", 0, "App instance index")
    app.argIP = flag.String("l", "0.0.0.0", "App local ip")
    app.argPort = flag.Int("p", 0, "App rpc port")
    //ptrWanIP = flag.String("w", "0.0.0.0", "App wan ip")
}

func (app *App) readArgs() {
    flag.Parse()
    if *app.argPort <= 0 {
        panic("no App port specified!")
    }
    if *app.argCluster <= 0 {
        panic("no App cluster id specified")
    }
    log.Printf("args: cluster<%v>, index<%v>, ip<%v>, port<%v>",
        *app.argCluster, *app.argIndex, *app.argIP, *app.argPort)
}

func (app *App) listenInterupt() {
    defer app.wg.Done()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
    app.shutdown()
}

func (app *App) shutdown() {
    app.logFunc("graceful shutdown...")
    close(app.done)
}

func (app *App) writePid() {
    pName := app.processName()
    pidFile := util.GenPidFilePath(pName)
    err := util.WritePidToFile(pidFile, os.Getpid())
    if err != nil {
        log.Print(err)
        return
    }
    log.Printf("write pid to %v", pidFile)
}

func (app *App) removePid() {
    pName := app.processName()
    pidFile := util.GenPidFilePath(pName)
    util.DeletePidFile(pidFile)
}

func (app *App) processName() string {
    if app.pName != "" {
        return app.pName
    }
    app.pName = filepath.Base(os.Args[0])
    if *app.argIndex > 0 {
        app.pName = fmt.Sprintf("%v%v", app.pName, *app.argIndex)
    }
    return app.pName
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
    return nil
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
    err := defaultApp.init()
    if err != nil {
        panic(err)
    }
    defaultApp.run()
}

// 获取server进程名字
func ProcessName() string {
    return defaultApp.processName()
}
