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
  
func renderPingPlot(pingResults *[]pingResult) {
	chartminjs, err := res.ReadFile("chart/chart.min.js")
	if err != nil {
		log.Fatal(err)
	}

	chartpluginzoomminjs, err := res.ReadFile("chart/chartjs-plugin-zoom.min.js")
	if err != nil {
		log.Fatal(err)
	}
	
	template, err := template.ParseFS(res, "chart/main-template-load.html")
    if err != nil {
        log.Fatalln(err)
    }

	file, err := ioutil.TempFile(os.TempDir(), "ppingplot.*.html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	templateVars := struct{
		Chartminjs string
		Chartpluginzoomminjs string
		PingResults []pingResult
	}{
		Chartminjs: string(chartminjs),
		Chartpluginzoomminjs: string(chartpluginzoomminjs),
		PingResults: *pingResults,
	}

	err = template.Execute(file, templateVars)
	if err != nil {
		fmt.Println(err)
	}

	openbrowser(file.Name())
}
