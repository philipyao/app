package app

import (
    "fmt"
    "os"
    "os/signal"
    "flag"
    "sync"
    "time"
    "path/filepath"
    "syscall"

    "github.com/philipyao/toolbox/util"
    "github.com/philipyao/phttp"
    "math/rand"
)

type App struct {
    pName       string

    bInited     bool

    done        chan struct{}
    wg          sync.WaitGroup

    argCluster  *string
    argIndex    *int

    services    []Service
    fnArgOpts   []FnArgOption
    fnOpts      []FnOption

    option
}

var defaultApp = NewApp()

func NewApp() *App {
    app := &App{done: make(chan struct{})}
    return app
}

func (app *App) ReadArgs(opts ...FnArgOption) {
    //app.fnArgOpts = opts
    app.prepareArgs()
    for _, opt := range opts {
        opt()
    }
    app.readArgs()
}

func (app *App) Init(opts ...FnOption) error {
    if app.bInited {
        panic("already inited.")
    }

    rand.Seed(time.Now().UnixNano())

    for _, opt := range opts {
        opt(&app.option)
    }
    if app.logFunc == nil {
        app.logFunc = defaultLogFunc()
    }
    app.logFunc("app init...")

    app.bInited = true
    return nil
}

func (app *App) Run(svcs ...Service) error {
    if !app.bInited {
        panic("not inited")
    }
    app.logFunc("app run with %v service(s)...", len(svcs))

    app.services = svcs
    var err error
    for _, srv := range app.services {
        err = srv.OnInit()
        if err != nil {
            return err
        }
    }
    for _, srv := range app.services {
        err := srv.Serve()
        if err != nil {
            return err
        }
    }
    app.writePid()

    app.wg.Add(1)
    go app.listenInterupt()

    app.wg.Wait()

    app.logFunc("finalize...")
    for i := len(app.services) - 1; i >= 0; i-- {
        app.services[i].OnFini()
    }
    app.removePid()
    return nil
}

func (app *App) Cluster() string{
    if !app.bInited {
        panic("not inited")
    }
    return *app.argCluster
}

func (app *App) Index() int {
    if !app.bInited {
        panic("not inited")
    }
    return *app.argIndex
}

func (app *App) ProcessName() string {
    if app.pName != "" {
        return app.pName
    }
    app.pName = filepath.Base(os.Args[0])
    if *app.argIndex > 0 {
        app.pName = fmt.Sprintf("%v%v", app.pName, *app.argIndex)
    }
    return app.pName
}

//====================================

func (app *App) prepareArgs() {
    app.argCluster = flag.String("cluster", "", "app cluster")
    app.argIndex = flag.Int("index", 0, "app index")
}

func (app *App) readArgs() {
    flag.Parse()
    if *app.argCluster == "" {
        panic("no App cluster specified")
    }
    if *app.argIndex <= 0 {
        panic("no App index specified or invalid index")
    }
}

func (app *App) listenInterupt() {
    defer app.wg.Done()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
    app.shutdown()
    for _, srv := range app.services {
        srv.Close()
    }
}

func (app *App) shutdown() {
    app.logFunc("receive cmd: graceful shutdown.")
    close(app.done)
}

func (app *App) writePid() {
    pName := app.ProcessName()
    pidFile := util.GenPidFilePath(pName)
    err := util.WritePidToFile(pidFile, os.Getpid())
    if err != nil {
        app.logFunc(err.Error())
        return
    }
    app.logFunc("pid file <%v> writen.", pidFile)
}

func (app *App) removePid() {
    pName := app.ProcessName()
    pidFile := util.GenPidFilePath(pName)
    util.DeletePidFile(pidFile)
    app.logFunc("pid file removed.")
}

//=====================================================

// 可选：开启 rpc 服务
func ServeRpc(addr, registry string, rcvr interface{}, name string) Service {
    return newServiceRpc(
        defaultApp.Cluster(),
        defaultApp.Index(),
        rcvr,
        name,
        addr,
        registry,
    )
}

//可选：开启 http 服务
func ServeHttp(addr string, handler func(worker *phttp.HTTPWorker) error) Service {
    return newServiceHttp(addr, handler)
}

//1. app 读取命令行参数，可以自定义参数
func ReadArgs(opts ...FnArgOption) {
    defaultApp.ReadArgs(opts...)
}

//2. app 初始化
func Init(opts ...FnOption) error {
    return defaultApp.Init(opts...)
}

//3. app 运行入口函数
func Run(svcs ...Service) error {
    return defaultApp.Run(svcs...)
}

func Cluster() string {
    return defaultApp.Cluster()
}

func Index() int {
    return defaultApp.Index()
}

// 获取server进程名字
func ProcessName() string {
    return defaultApp.ProcessName()
}
