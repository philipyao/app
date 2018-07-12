package app

import (
    "testing"
    "os"
    "github.com/philipyao/phttp"
)

func TestRunApp(t *testing.T) {
    os.Args = []string{"appTest", "-cluster", "world100", "-index", "1"}
    ReadArgs()
    err := Init()
    if err != nil {
        t.Fatal(err)
    }

    err = Run(
        ServeHttp(":12003", func(w *phttp.HTTPWorker) error {
            return nil
        }),
    )
    if err != nil {
        t.Fatal(err)
    }
}