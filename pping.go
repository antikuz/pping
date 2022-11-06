package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type pingResult struct {
	PingTime time.Time
	Latency  int
}

type pingStatistic struct {
	Min         int
	Max         int
	Transmitted int
	Received    int
}

var (
	t = flag.Bool("t", false, "Ping the specified host until stopped. To stop - type Control-C.")
	n = flag.Int("n", 4, "Number of echo requests to send.")
	w = flag.String("w", "1000", "Timeout in milliseconds to wait for each reply.")

	pingResults = &[]pingResult{}
	pingStatistics = &pingStatistic{}
)

func ping(destination string) (int, error) {
	var stdout []byte
	var err error

	if runtime.GOOS == "windows" {
		stdout, err = exec.Command("ping", "-n", "1", "-w", *w, destination).CombinedOutput()
	} else {
		stdout, err = exec.Command("ping", "-w", "1", "-W", *w, destination).CombinedOutput()
	}

	if err != nil {
		return 0, fmt.Errorf("%v: %s", err, string(stdout))
	}

	re, err := regexp.Compile(`time[=<](\d)`)
	if err != nil {
		return 0, err
	}

	res := re.FindSubmatch(stdout)
	if res == nil {
		return 0, nil
	}

	if string(res[1]) == "" {
		err = fmt.Errorf("%s", string(stdout))
		return 0, err
	}
	latency, err := strconv.Atoi(string(res[1]))
	if err != nil {
		return 0, err
	}

	return latency, nil
}

func pingResultContainError(err error) bool {
	switch {
	case strings.Contains(err.Error(), "timed out"):
		return false
	case strings.Contains(err.Error(), "host unreachable"):
		return false
	case strings.Contains(err.Error(), "0 received"):
		return false
	default:
		return true
	}
}

func pingStatisticUpdate(ps *pingStatistic, result int) {
	ps.Transmitted += 1
	if result == -1 {
		return
	}
	ps.Received += 1
	if result < ps.Min {
		ps.Min = result
	}
	if result > ps.Max {
		ps.Max = result
	}
}

func pingStatisticLine(ps *pingStatistic) string {
	packetLoss := (float64(ps.Transmitted - ps.Received)/float64(ps.Transmitted)) * 100
	return fmt.Sprintf("%d packets transmitted, %d received, %.0f%% packet loss", ps.Transmitted, ps.Received, packetLoss)
}

func main() {
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

	defer renderPingChart(pingResults, pingStatistics, destination)

	go func() {
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
						if pingResultContainError(err) {
							log.Fatal(err)
						}
						result = -1
					}

					*pingResults = append(*pingResults, pingResult{
						PingTime: time.Now(),
						Latency:  result,
					})
					pingStatisticUpdate(pingStatistics, result)
					if result == -1 {
						log.Printf("Request timed out.%s", strings.Repeat(" ", 60))
						fmt.Printf("%s\r", pingStatisticLine(pingStatistics))
					} else {
						log.Printf("time=%dms%s", result, strings.Repeat(" ", 60))
						fmt.Printf("%s\r", pingStatisticLine(pingStatistics))
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	} else {
		for i := *n; i > 0; i-- {
			select {
			case <-ticker.C:
				wg.Add(1)
				go func() {
					result, err := ping(destination)
					if err != nil {
						if pingResultContainError(err) {
							log.Fatal(err)
						}
						result = -1
					}

					*pingResults = append(*pingResults, pingResult{
						PingTime: time.Now(),
						Latency:  result,
					})

					if result == -1 {
						log.Println("Request timed out.")
					} else {
						log.Printf("time=%dms", result)
					}
					wg.Done()
				}()
			case <-ctx.Done():
				return
			}
		}
		wg.Wait()
	}
}
