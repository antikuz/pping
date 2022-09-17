package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

func ping(destination string) (string, error) {
	stdout, err := exec.Command("ping", "-n", "1", destination).CombinedOutput()
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
	count := flag.Int("n", 1, "count")
	flag.Parse()

	destination := flag.Arg(0)
	if destination == "" {
		flag.Usage()
		os.Exit(1)
	}

	for i := *count; i > 0; i-- {
		result, err := ping(destination)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("time=%s\n", result)
		if i != 1 {
			time.Sleep(time.Second)
		}
	}
}
