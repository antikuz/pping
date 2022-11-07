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

type regexpResult struct {
	target string
	bytes  int
	time   int
	ttl    int
}

var (
	t = flag.Bool("t", false, "Ping the specified host until stopped. To stop - type Control-C.")
	n = flag.Int("n", 4, "Number of echo requests to send.")
	s = flag.Int("s", 32, "Send buffer size.")
	w = flag.String("w", "1000", "Timeout in milliseconds to wait for each reply.")
	g = flag.Bool("g", false, "Generate web graph after exit.")

	pingResults    = &[]pingResult{}
	pingStatistics = &pingStatistic{}
	re             *regexp.Regexp
)

func ping(destination string) (*regexpResult, error) {
	var stdout []byte
	var err error

	if runtime.GOOS == "windows" {
		size := strconv.Itoa(*s)
		stdout, err = exec.Command("ping", "-n", "1", "-w", *w, "-l", size, destination).CombinedOutput()
	} else {
		size := strconv.Itoa(*s - 8)
		stdout, err = exec.Command("ping", "-n", "-w", "1", "-W", *w, "-s", size, destination).CombinedOutput()
	}

	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, string(stdout))
	}

	match := re.FindStringSubmatch(string(stdout))
	regexpMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			regexpMap[name] = match[i]
		}
	}

	bytes, err := strconv.Atoi(regexpMap["bytes"])
	if err != nil {
		return nil, err
	}

	time, err := strconv.Atoi(regexpMap["time"])
	if err != nil {
		return nil, err
	}

	ttl, err := strconv.Atoi(regexpMap["ttl"])
	if err != nil {
		return nil, err
	}

	reResult := &regexpResult{
		target: regexpMap["target"],
		bytes:  bytes,
		time:   time,
		ttl:    ttl,
	}

	return reResult, nil
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
	packetLoss := (float64(ps.Transmitted-ps.Received) / float64(ps.Transmitted)) * 100
	return fmt.Sprintf("%d packets transmitted, %d received, %.0f%% packet loss", ps.Transmitted, ps.Received, packetLoss)
}

func pingResultProcessing(result regexpResult, err error) {
	if err != nil {
		if pingResultContainError(err) {
			log.Fatal(err)
		}
		result.time = -1
	}

	*pingResults = append(*pingResults, pingResult{
		PingTime: time.Now(),
		Latency:  result.time,
	})
	pingStatisticUpdate(pingStatistics, result.time)
	if result.time == -1 {
		log.Printf("Request timed out.%s", strings.Repeat(" ", 60))
		fmt.Printf("%s\r", pingStatisticLine(pingStatistics))
	} else {
		log.Printf("Reply from %s: bytes=%d time=%dms TTL=%d%s", result.target, result.bytes, result.time, result.ttl, strings.Repeat(" ", 60))
		fmt.Printf("%s\r", pingStatisticLine(pingStatistics))
	}
}

func main() {
	flag.Usage = func() {
		fmt.Print(`Usage of pping:
    -t             Ping the specified host until stopped. To stop - type Control-C.
    -n count       Number of echo requests to send.
    -s size        Send buffer size.
    -w timeout     Timeout in milliseconds to wait for each reply.
    -g graph       Generate web graph after exit.
    `)
		os.Exit(0)
	}
	flag.Parse()

	destination := flag.Arg(0)
	if destination == "" {
		flag.Usage()
	}

	if runtime.GOOS == "windows" {
		re = regexp.MustCompile(`from (?P<target>.*): bytes=(?P<bytes>\d+) time=(?P<time>\d+).*TTL=(?P<ttl>\d+)`)
	} else {
		re = regexp.MustCompile(`(?P<bytes>\d+) bytes from (?P<target>.*):.*ttl=(?P<ttl>\d+) time=(?P<time>\d+)`)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	ticker := time.NewTicker(1 * time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	if *g {
		defer renderPingChart(pingResults, pingStatistics, destination)
	}

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
					pingResultProcessing(*result, err)
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
					pingResultProcessing(*result, err)
					wg.Done()
				}()
			case <-ctx.Done():
				return
			}
		}
		wg.Wait()
	}
}
