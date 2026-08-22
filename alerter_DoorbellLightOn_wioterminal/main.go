package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"machine"
	"strings"
	"time"

	"github.com/sago35/tinygo-examples/wioterminal/initialize"
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/net"
)

//go:embed gotify_logo.jpg
var gotify_image []byte
var display *ili9341.Device

var (
	ssid     string
	password string
)

func main() {
	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

	ssid = "<YOUR WIFI SSID>"
	password = "<YOUR WIFI PASSWORD>"

	// WioTerminalの左ボタンを押すと液晶バックライトOFFにする
	button1 := machine.PC28 // WioTerminal Button3 Left
	button1.Configure(machine.PinConfig{Mode: machine.PinInput})
	button1.SetInterrupt(machine.PinToggle, func(machine.Pin) {
		machine.LCD_BACKLIGHT.Low() // 液晶のバックライトOFF
	})

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
	led.High()

	display = initdisplay.InitDisplay() // ここで液晶のバックライトがONになる machine.LCD_BACKLIGHT.High()が呼ばれている
	img, err := jpeg.Decode(strings.NewReader(string(gotify_image)))
	if err != nil {
		log.Fatal(err)
	}
	display.FillScreen(white)

	time.Sleep(2 * time.Second)
	machine.LCD_BACKLIGHT.Low() // 液晶のバックライトOFF

	err = run(led, img)
	if err != nil {
		log.Fatal(err)
	}

	select {}

}

var (
	port int
)

func run(led machine.Pin, img image.Image) error {
	port = 4416 // udp port

	ip := net.ParseIP("255.255.255.255") // broadcast
	raddr := &net.UDPAddr{IP: ip, Port: port}
	laddr := &net.UDPAddr{Port: port}

	conn, err := net.DialUDP("udp", laddr, raddr)
	for ; err != nil; conn, err = net.DialUDP("udp", laddr, raddr) {
		time.Sleep(5 * time.Second)
	}

	fmt.Printf("UDP listen : %d\r\n", port)
	buf := [32]byte{}
	n := int(0)
	for {
		// バッファがあるためか、ある程度データが蓄積されないと読み込みされない?
		// 送信間隔を1分くらい空けないと、次のパケットを読みとらない。
		// 同じデータを送信しても反応しないことが多い。
		n, err = conn.Read(buf[:])
		if err != nil {
			fmt.Println("conn.Read() error!")
			// return err
		} else if n >= 5 {
			// fmt.Printf("recv: %s\r\n", string(buf[:n]))
			recv_msg := string(buf[:5])
			//			fmt.Printf("recv: [%s]\r\n", recv_msg)
			if recv_msg == "Alert" {
				fmt.Println("recv Alert")

				led.Low()
				time.Sleep(100 * time.Millisecond)
				led.High()
				time.Sleep(100 * time.Millisecond)
				led.Low()
				time.Sleep(100 * time.Millisecond)
				led.High()

				displayImage(img)
				machine.LCD_BACKLIGHT.High() // 液晶のバックライトON
			}
		}
		time.Sleep(1 * time.Second)
	}
	conn.Close()
	return nil
}

func displayImage(img image.Image) {

	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			display.SetPixel(int16(x), int16(y), color.RGBA{
				R: uint8(r >> 8), G: uint8(g >> 8),
				B: uint8(b >> 8), A: uint8(0xFF)})
		}
	}
}
