package app

import (
    "testing"
    "fmt"
    "os"
    "log"
)

func TestRunApp(t *testing.T) {
    // cluster: 100
    // index: 1
    os.Args = []string{"appTest", "-c", "world100", "-i", "1"}
    var err error
    err = UseInit(
        func() error {
            fmt.Println("== init ok.")
            return nil
        })
    if err != nil {
        log.Fatalf("srv.HandleBase() err: %v", err)
    }
    Run()
}