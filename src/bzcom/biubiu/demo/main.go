package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
  "time"
)

func catchSignal() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGUSR1)
	<-sig
	log.Println("recieved signal!")
	panic("")
}

type testDemo struct{
  value int
  data []byte
}

// this is for demo
func main() {

	go catchSignal()

	go func() {
		log.Println(http.ListenAndServe(":8089", nil))
	}()

  lmap := make(map[int]*testDemo)
  for i :=0 ;i  < 100000;i++{
    lmap[i] = &testDemo{
      value:i,
      data:make([]byte,1),
    }
  }

  for{
    time.Sleep(time.Millisecond*50)
    for k,v:=range lmap{
      if k > 1000000{
        log.Println("v",v)
      }
    }
  }

	select {}
}
