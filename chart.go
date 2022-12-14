package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"text/template"
)

var (
	//go:embed chart
	res embed.FS
)

func openbrowser(file string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", file).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", file).Start()
	case "darwin":
		err = exec.Command("open", file).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func renderPingChart(pingResults *[]pingResult, ps *pingStatistic, destination string) {
	chartminjs, err := res.ReadFile("chart/chart.min.js")
	if err != nil {
		log.Fatal(err)
	}

	chartjsadapterdatefnsbundleminjs, err := res.ReadFile("chart/chartjs-adapter-date-fns.bundle.min.js")
	if err != nil {
		log.Fatal(err)
	}

	hammerjs, err := res.ReadFile("chart/hammerjs@2.0.8.js")
	if err != nil {
		log.Fatal(err)
	}

	chartpluginzoomminjs, err := res.ReadFile("chart/chartjs-plugin-zoom.min.js")
	if err != nil {
		log.Fatal(err)
	}

	template, err := template.ParseFS(res, "chart/chart-template.html")
	if err != nil {
		log.Fatalln(err)
	}

	file, err := ioutil.TempFile(os.TempDir(), "ppingplot.*.html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	templateVars := struct {
		Chartminjs                       string
		Chartjsadapterdatefnsbundleminjs string
		Hammerjs                         string
		Chartpluginzoomminjs             string
		PingResults                      []pingResult
		PingStatistic                    pingStatistic
		Destination                      string
	}{
		Chartminjs:                       string(chartminjs),
		Chartjsadapterdatefnsbundleminjs: string(chartjsadapterdatefnsbundleminjs),
		Hammerjs:                         string(hammerjs),
		Chartpluginzoomminjs:             string(chartpluginzoomminjs),
		PingResults:                      *pingResults,
		PingStatistic:                    *ps,
		Destination:                      destination,
	}

	err = template.Execute(file, templateVars)
	if err != nil {
		fmt.Println(err)
	}

	openbrowser(file.Name())
}
