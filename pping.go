package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"time"
)

type pingResult struct {
	PingTime time.Time
	Latency string
}

var pingResults = &[]pingResult{}

func ping(destination string) (string, error) {
	var stdout []byte
	var err error
	
	if runtime.GOOS == "windows" {
		stdout, err = exec.Command("ping", "-n", "1", destination).CombinedOutput()
	} else {
		stdout, err = exec.Command("ping", "-w", "1", destination).CombinedOutput()
	}

	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(`time=(\d)`)
	if err != nil {
		return "", err
	}
	res := re.FindSubmatch(stdout)
	if res == nil {
		return "", nil
	}

	return string(res[1]), nil
}

func main() {
	count := flag.Int("n", 4, "count")
	t := flag.Bool("t", false, `Ping the specified host until stopped.
To stop - type Control-C.`)
	
	flag.Parse()

	destination := flag.Arg(0)
	if destination == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	ticker := time.NewTicker(1 * time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	
	defer renderPingChart(pingResults)
	
	go func(){
		for range c {
			cancel()
		}
	}()
	
	if *t {
		for {
			select {
			case <-ticker.C:
				go func() {
					result, err := ping(destination)
					if err != nil {
						log.Fatal(err)
					}
					
					*pingResults = append(*pingResults, pingResult{
						PingTime: time.Now(),
						Latency: result,
					})
					log.Printf("time=%sms\n", result)
				}()
			case <-ctx.Done():
				return
			}
		}
	} else {
		for i := *count; i > 0; i-- {
			select {
			case <-ticker.C:
				wg.Add(1)
				go func() {
					result, err := ping(destination)
					if err != nil {
						log.Fatal(err)
					}
					*pingResults = append(*pingResults, pingResult{
						PingTime: time.Now(),
						Latency: result,
					})
					log.Printf("time=%sms\n", result)
					wg.Done()
				}()
			case <-ctx.Done():
				return
			}
		}
		wg.Wait()
	}
}