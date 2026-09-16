package main

import (
	"fmt"
	"log"
	"machine"
	"net"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid         string
	password     string
	alerter_addr string // ex: "<alerter_ip>:4416"
)

func main() {
	ssid = "<Your WiFi ssid>"
	password = "<Your WiFi passphrase>"

	machine.InitADC()
	sensor := machine.ADC{Pin: machine.WIO_LIGHT} // wio terminal photo detector
	//	sensor.Configure(machine.ADCConfig{Resolution: 12})
	sensor.Configure(machine.ADCConfig{})

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led.Low()
	time.Sleep(2 * time.Second)

	fmt.Println("WiFi Connect start")

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: password,
	})
	if err != nil {
		fmt.Println("WiFi Connect fail!")
		log.Fatal(err)
	}

	fmt.Println("WiFi Connect success")

	masktime := 0
	cntr := 0
	for {
		// 一度、検出したら、一定時間反応しないようにする必要がある。( 300 = 60sec * 5min)
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
			fmt.Println("detect LightOn!")
			masktime = 300 //  (300 = 60sec * 5min)

			err = run()
			if err != nil {
				// log.Fatal(err)
				fmt.Println(err)
			}

		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}
	link.NetDisconnect()

}

func run() error {
	alerter_addr = "<Your alerter_device IPaddress>:<port>"

	conn, err := net.Dial("tcp", alerter_addr)
	for ; err != nil; conn, err = net.Dial("tcp", alerter_addr) {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
	}

	for i := 0; i < 1; i++ {
		fmt.Printf("write to %s\n", alerter_addr)
		//fmt.Fprintf(conn, "Alert")
		conn.Write([]byte("Alert"))
		time.Sleep(500 * time.Millisecond)
	}
	conn.Close()

	return nil
}
