package main

import (
	"encoding/json"
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

type Sensor struct {
	PM2_5        int `json:"pm25"`
	Formaldehyde int `json:"formaldehyde"`
	CO2          int `json:"co2"`
	Temperature  int `json:"temperature"`
	Humidity     int `json:"humidity"`
}

var sensor = Sensor{}

var gauges = [][]string{
	[]string{"unknown", "unknown"},
	[]string{"pm2dot5", "PM 2.5"},
	[]string{"formaldehyde", "Formaldehyde"},
	[]string{"co2", "CO2"},
	[]string{"temperature", "Temperature"},
	[]string{"humidity", "Humidity"},
	[]string{"voc", "VOC"},
}

var gaugesArray = []prometheus.Gauge{}

func setupGauges() {
	gaugesArray = []prometheus.Gauge{}
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

	debugInfo := []string{}

	for i := 0; i < 7; i++ {
		if value, err := strconv.Atoi(items[i]); err == nil {
			gaugesArray[i].Set(float64(value))
			if *debug {
				gauge := gauges[i]
				debugInfo = append(debugInfo, fmt.Sprintf("%v: %v", gauge[1], value))
			}

			switch i {
			case 1:
				sensor.PM2_5 = value
			case 2:
				sensor.Formaldehyde = value
			case 3:
				sensor.CO2 = value
			case 4:
				sensor.Temperature = value
			case 5:
				sensor.Humidity = value
			}
		}

	}
	log.Println(strings.Join(debugInfo, ","))
}

func listenOnSerialPort() {
	c := &serial.Config{Name: *serialPortFile, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Reading device data every 30 seconds...")

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

	h := func(w http.ResponseWriter, _ *http.Request) {
		b, err := json.Marshal(sensor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}

	log.Println("Prometheus metrics listens on", *addr)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/json", h)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
