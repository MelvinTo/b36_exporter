## What is it
* A tool to programtically read the live data from air quality measurement device B36 and provide exporter api for prometheus server to read.
* So in the end, you can use prometheus+grafana to draw charts on air quality data
* This tool can only run on Linux so far

## How to build
* Install latest golang
* Build
```
go build
```
## How to use
* Connect the B36 device to the USB port of a Linux machine
* Run this command on Linux
```
# need root priviledge to read serial port
sudo ./b36_exporter
```
* Help
```
$ sudo ./b36_exporter --help 
Usage of ./b36_exporter:
  -debug
        whether to print raw data on console
  -listen-address string
        The address to listen on for prometheus request. (default ":9301")
  -serial string
        the file path of the serial port (default "/dev/ttyUSB0")
```
