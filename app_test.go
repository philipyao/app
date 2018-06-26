package app

import (
    "testing"
    "os"
    "github.com/philipyao/phttp"
)

func TestRunApp(t *testing.T) {
    os.Args = []string{"appTest", "-c", "world100", "-i", "1"}
    UseServiceHttp(":12003", func(w *phttp.HTTPWorker) error {
        return nil
    })
    Run()
}