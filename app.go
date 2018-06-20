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
)

type App struct {
    pName       string

    bInited     bool

    done        chan struct{}
    wg          sync.WaitGroup

    argCluster  *string
    argIndex    *int

    fnInit      func() error
    fnServe     func(<-chan struct{}) error
    fnFini      func()
    logFunc     func(format string, args ...interface{})
}

var defaultApp  = &App{done: make(chan struct{})}
func init() {
    defaultApp.prepare()
}

func (app *App) prepare() {
    app.prepareArgs()
}

func (app *App) Init() error {
    if app.logFunc == nil {
        app.logFunc = defaultLogFunc()
    }

    app.logFunc("init...")

    if app.bInited {
        panic("already inited.")
    }

    app.readArgs()

    var err error
    if app.fnInit != nil {
        err = app.fnInit()
        if err != nil {
            return err
        }
    }

    app.bInited = true
    app.logFunc("init ok.")
    return nil
}

func (app *App) Run() {
    app.logFunc("run...")
    if !app.bInited {
        panic("not inited")
    }
    if app.fnServe != nil {
        err := app.fnServe(app.done)
        if err != nil {
            app.logFunc(err.Error())
            return
        }
    }
    app.writePid()

    app.wg.Add(1)
    go app.listenInterupt()

    app.wg.Wait()

    app.logFunc("finalize...")
    if app.fnFini != nil {
        app.fnFini()
    }
    app.removePid()
}

//====================================

func (app *App) prepareArgs() {
    app.argCluster = flag.String("c", "", "App cluster")
    app.argIndex = flag.Int("i", 0, "App index")
}

func (app *App) readArgs() {
    flag.Parse()
    if *app.argCluster == "" {
        panic("no App cluster specified")
    }
    if *app.argIndex <= 0 {
        panic("no App index specified or invalid index")
    }
    app.logFunc("args: cluster<%v>, name<%v>", *app.argCluster, app.processName())
}

func (app *App) listenInterupt() {
    defer app.wg.Done()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
    app.shutdown()
}

func (app *App) shutdown() {
    app.logFunc("receive cmd: graceful shutdown.")
    close(app.done)
}

func (app *App) writePid() {
    pName := app.processName()
    pidFile := util.GenPidFilePath(pName)
    err := util.WritePidToFile(pidFile, os.Getpid())
    if err != nil {
        app.logFunc(err.Error())
        return
    }
    app.logFunc("pid file <%v> writen.", pidFile)
}

func (app *App) removePid() {
    pName := app.processName()
    pidFile := util.GenPidFilePath(pName)
    util.DeletePidFile(pidFile)
    app.logFunc("pid file removed.")
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

// app 自定义初始化
func UseInit(fnInit func() error) error {
    if fnInit == nil {
        return errors.New("nil fnInit")
    }
    defaultApp.fnInit = fnInit
    return nil
}

// app 自定义服务
func UseServe(fnServe func(<-chan struct{}) error) error {
    if fnServe == nil {
        return errors.New("nil fnServe")
    }
    defaultApp.fnServe = fnServe
    return nil
}

// app 自定义回收
func UseFini(fnFini func()) error {
    if fnFini == nil {
        return errors.New("nil fnFini")
    }
    defaultApp.fnFini = fnFini
    return nil
}

// app 运行入口函数
func Run() {
    err := defaultApp.Init()
    if err != nil {
        panic(err)
    }
    defaultApp.Run()
}

//可选，自定义log输出
func SetLogger(l func(int, string, ...interface{})) {
    defaultApp.logFunc = customLogFunc(l)
}


// 获取server进程名字
func ProcessName() string {
    return defaultApp.processName()
}
