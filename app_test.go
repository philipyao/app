package app

import (
    "testing"
    "log"
    "os"
)

func TestRunApp(t *testing.T) {
    os.Args = []string{"appTest", "-c", "100", "-i", "1", "-l", "127.0.0.1", "-p", "10021"}
    var err error
    err = HandleBase(
        func(done chan struct{}) error {
            log.Println("app init ok.")
            return nil
        },
        func () {
            log.Println("app shutdown ok.")
        },
    )
    if err != nil {
        log.Fatalf("srv.HandleBase() err: %v", err)
    }
    Run()
}