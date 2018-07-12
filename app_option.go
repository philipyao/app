package app

import "flag"

type option struct {
    logFunc     func(format string, args ...interface{})
}
type FnOption func(opt *option)
type FnArgOption func()

//======================= arg options ========================
func WithArgInt(p *int, name string, value int, usage string) FnArgOption {
    return func() {
        flag.IntVar(p, name, value, usage)
    }
}
func WithArgString(p *string, name string, value string, usage string) FnArgOption {
    return func() {
        flag.StringVar(p, name, value, usage)
    }
}
func WithArgBool(p *bool, name string, value bool, usage string) FnArgOption {
    return func() {
        flag.BoolVar(p, name, value, usage)
    }
}

//======================= arg options ========================
func WithLogger(logFunc func(string, ...interface{})) FnOption {
    return func(opt *option) {
        opt.logFunc = logFunc
    }
}

//todo
func WithSignalHandle() FnOption {
    return func(opt *option) {

    }
}

//todo
func WithPprof() FnOption {
    return func(opt *option) {

    }
}

//todo
func WithCPUNum(num int) FnOption {
    return func(opt *option) {

    }
}

//todo
func WithReload() FnOption {
    return func(opt *option) {

    }
}