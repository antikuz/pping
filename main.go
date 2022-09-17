package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
)

func main(){
	flag.String("dest", "", "Ping destination")
	flag.Parse()

	destination := flag.Arg(0)
	if destination == "" {
		flag.Usage()
		os.Exit(1)
	}

    cmd := exec.Command("ping", destination)
	cmd.StdoutPipe()
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}