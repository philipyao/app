package app

import (
    "testing"
    "log"
)

func TestRunApp(t *testing.T) {
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