package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"machine"
	"net"
	"os"
	"strings"
	"time"

	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

//go:embed gotify_logo.jpg
var gotify_image []byte
var display *ili9341.Device

var (
	ssid     string
	password string
	port     int
)

func main() {
	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

	ssid = "<Your WiFi ssid>"
	password = "<Your WiFi passphrase>"

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

	led.High()

	display = InitDisplay() // 液晶ディスプレイ初期化

	img, err := jpeg.Decode(strings.NewReader(string(gotify_image)))
	if err != nil {
		log.Fatal(err)
	}
	display.FillScreen(white)
	time.Sleep(2 * time.Second)
	machine.LCD_BACKLIGHT.Low() // 液晶のバックライトOFF

	port = 4416
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("TCP server is running on %s\n", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}

		go run(conn, led, img)
	}

}

func run(conn net.Conn, led machine.Pin, img image.Image) error {
	defer conn.Close()

	fmt.Printf("Client Connected:", conn.RemoteAddr())

	buf := [1024]byte{}
	for {
		n, err := conn.Read(buf[:])
		if err != nil && err == io.EOF {
			fmt.Println(err)
			break
		} else if err != nil {
			fmt.Println(err)
			continue
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
	fmt.Println("Client Disconnected.")

	return nil
}

// "tinygo.org/x/drivers/examples/ili9341/initdisplay"
func InitDisplay() *ili9341.Device {

	backlight := machine.LCD_BACKLIGHT
	backlight.Configure(machine.PinConfig{Mode: machine.PinOutput})

	spi := machine.SPI3
	spi.Configure(machine.SPIConfig{
		Frequency: 40000000,
		SCK:       machine.LCD_SCK_PIN,
		SDO:       machine.LCD_SDO_PIN,
		SDI:       machine.LCD_SDI_PIN,
	})
	//	cs := machine.LCD_SCK_PIN
	/*
		cs := machine.LCD_SS_PIN
		dc := machine.LCD_DC
		rst := machine.LCD_RESET

		cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
		dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
		rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	*/

	display := ili9341.NewSPI(
		spi,
		machine.LCD_DC,
		machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)

	display.Configure(ili9341.Config{})

	backlight.High()

	display.SetRotation(ili9341.Rotation270)

	return display
}

func displayImage(img image.Image) {

	//	fmt.Printf("Y Size=%d, X Size=%d\r\n", img.Bounds().Max.Y, img.Bounds().Max.X)

	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			display.SetPixel(int16(x), int16(y), color.RGBA{
				R: uint8(r >> 8), G: uint8(g >> 8),
				B: uint8(b >> 8), A: uint8(0xFF)})
		}
	}
}
