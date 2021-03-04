package main

import (
	"context"
	"log"
	"merpc"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr chan string) {
	var foo Foo

	//pick a free port
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("network error:", err)
	}

	if err := merpc.DefaultServer.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}

	log.Println("start rpc server on", l.Addr())
	merpc.HandleHTTP()
	addr <- l.Addr().String()
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	client, _ := merpc.DialHTTP("tcp", <-addrCh)
	defer func() { _ = client.Close() }()
	time.Sleep(time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i} //fmt.Sprintf("merpc req %d", i)
			var reply int                       //string
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error - ", err)
			}
			log.Printf("%d + %d = %d\n", args.Num1, args.Num2, reply)
		}(i)

	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	startServer(addr)
}
