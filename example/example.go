package main

import (
	wgt "github.com/dmdv/waitgrouptimeout"
	"time"
)

func main() {
	wg := wgt.New(true)
	wg.Wrap(func() {
		time.Sleep(5 * time.Second)
		println("Hello, world 1!")
	})

	wg.Wrap(func() {
		time.Sleep(5 * time.Second)
		println("Hello, world 2!")
	})

	wg.Start()
	wg.WaitTimeout(2 * time.Second)
	if wg.Finished() {
		println("Finished")
	} else {
		println("Not finished")
	}

	for wg.Finished() != true {
		time.Sleep(1 * time.Second)
		println("Waiting...")
	}
	wg.Wait()
}
