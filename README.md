# App
app is a server app framework for game development

## Feature
* managed by cluster id and app index, which is common in game development
* graceful startup and shutdown
* support for rpc service and http serving
* use pid which is helpful for monitoring
 
 ## Get started
 
 ```golang
 
 import (
    "log"
    "github.com/philipyao/app"
 )
 
 func main() {
     var err error
     err = app.HandleBase(
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
     app.Run()
 }
     
 ```    
 
 After successful build, you will get an binary app, say appTest
 then just start it with some arguments like this:
 ./appTest -c 100 -i 1 -l "127.0.0.1" -p 10021
 
 ```
 
 2018/06/03 11:29:21 [srv]App start...
 2018/06/03 11:29:21 args: cluster<100>, index<1>, ip<127.0.0.1>, port<10021>
 2018/06/03 11:29:21 app init ok.
 2018/06/03 11:29:21 [srv]App init ok.
 2018/06/03 11:29:21 write pid to /Users/philip/work/git/src/github.com/philipyao/app/pid/run.appTest1.pid
 
 ```
 
 if you want to stop it, just CTRL + C, and it will graceful shutdown
 
 ```
 
 ^C2018/06/03 11:29:49 [srv]graceful shutdown...
 2018/06/03 11:29:49 app shutdown ok.
 
 ```