package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tarm/serial"
)

var serialPortFile = flag.String("serial", "/dev/ttyUSB0", "the file path of the serial port")
var baud = 57600
var addr = flag.String("listen-address", ":9301", "The address to listen on for prometheus request.")
var debug = flag.Bool("debug", false, "whether to print raw data on console")

var gaugesArray = []prometheus.Gauge{}

func setupGauges() {
	gauges := [][]string{
		[]string{"unknown", "unknown"},
		[]string{"pm2dot5", "PM 2.5"},
		[]string{"formaldehyde", "Formaldehyde"},
		[]string{"co2", "CO2"},
		[]string{"temperature", "Temperature"},
		[]string{"humidity", "Humidity"},
		[]string{"voc", "VOC"},
	}

	for _, g := range gauges {
		key := g[0]
		desc := g[1]
		gaugesArray = append(gaugesArray, promauto.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("b36_gauge_%v", key),
			Help: fmt.Sprintf("b36 gauge for %v", desc),
		}))
	}
}

func processData(input string) {
	input = strings.TrimSuffix(input, "\n")
	items := strings.Split(input, ",")
	if len(items) != 8 {
		if *debug {
			log.Println("Skipped invalid input:", input)
		}
		return
	}

	if *debug {
		log.Println(input)
	}

	for i := 0; i < 7; i++ {
		if value, err := strconv.Atoi(items[i]); err == nil {
			gaugesArray[i].Set(float64(value))
		}
	}
}

func listenOnSerialPort() {
	c := &serial.Config{Name: *serialPortFile, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Reading...")

	for {
		buf := make([]byte, 128)
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		str := string(buf[:n])
		for _, s := range strings.Split(str, "\n") {
			if s != "" {
				processData(s)
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func main() {
	flag.Parse()
	setupGauges()

	go listenOnSerialPort()

	log.Println("prometheus metrics listens on", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
