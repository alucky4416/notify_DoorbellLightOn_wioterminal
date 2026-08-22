package main

import (
	"fmt"
	"io"
	"log"
	"machine"
	"net"
	"strings"
	"time"

	"github.com/sago35/tinygo-examples/wioterminal/initialize"
	"tinygo.org/x/drivers/net/http"
)

var (
	ssid     string
	password string
)

func main() {
	ssid = "<YOUR WIFI SSID>"
	password = "<YOUR WIFI PASSWORD>"

	machine.InitADC()
	sensor := machine.ADC{Pin: machine.WIO_LIGHT} // wio terminal photo detector
	//	sensor.Configure(machine.ADCConfig{Resolution: 12})
	sensor.Configure(machine.ADCConfig{})

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led.Low()
	time.Sleep(2 * time.Second)

	fmt.Println("WiFi Connect start")

	_, err := initialize.Wifi(ssid, password, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("WiFi Connect success")

	masktime := 0
	cntr := 0
	for {
		// 一度、検出したら、一定時間反応しないようにする ( 300 = 60sec * 5min)
		if masktime > 0 {
			masktime--
			// fmt.Printf("masktime = %d\n", masktime)
			time.Sleep(1 * time.Second)
			continue
		}
		masktime = 0

		adcdata := sensor.Get()
		fmt.Printf("%04x\n", adcdata)

		if adcdata > 0x0500 {
			cntr++
		} else {
			cntr = 0
			led.Low()
		}
		if cntr > 4 { // 5回連続(1秒以上)閾値越えしたら、通知送信 200ms * 5 = 1sec
			cntr = 0
			led.High()
			fmt.Println("over time!")
			masktime = 300 //  (300 = 60sec * 5min)

			// send notify to Alerter 
			err = run()
			if err != nil {
				log.Fatal(err)
			}

		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func run() error {
	var port int

	port = 4416
	conn, _ := net.Dial("udp", fmt.Sprintf("255.255.255.255:%d", port))
	for i := 0; i < 3; i++ {
		fmt.Printf("send to 255.255.255.255:%d", port)
		fmt.Fprintf(conn, "Alert")
		time.Sleep(500 * time.Millisecond)
	}
	conn.Close()

	return nil
}
