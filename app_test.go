package app

import (
    "testing"
    "fmt"
    "os"
    "log"
)

func TestRunApp(t *testing.T) {
    os.Args = []string{"appTest", "-c", "world100", "-i", "1"}
    var err error
    err = UseInit(
        func() error {
            fmt.Println("== init ok.")
            return nil
        })
    if err != nil {
        log.Fatalf("srv.UseInit() err: %v", err)
    }
    Run()
}