notify_doorbell_light_wioterminal
===== 

# 概要
2台のWioTerminalを使って、
1台めのWioTerminal背面光センサーで室内側ドアホン液晶画面画面の点灯を検知して通知、
2台めのWioTerminalがその通知を受けて、WioTerminal液晶画面に通知が来たことを知らせる画像を表示する仕組み。

# 必要なもの
基本的にマンションなどのインターホン付きの個人宅での使用を想定。
玄関側インターホンから呼び出しボタンを押すと、宅内のインターホン液晶画面が点灯するようになっている必要がある。
また、宅内WiFiで2台のWioTerminalが通信できるようになっていること。

# 取り付け
## 1.1台目のWioTerminal (インターホン液晶の光検出)
　光センサーはWioTerminal背面側にあるので、宅内インターホンの液晶画面にWioTerminal背面があたるようにして取り付ける。
WioTerminalの電源は、USB充電アダプタなどを使う。　　

## 2.2台目のWioTerminalを電源ON
　2台目はWiFiが届くところならどこでも設置可能。電源は同じくUSB充電アダプタなどを使う。
1台目からの通知は、UDPブロードキャストで送信される。これを受信すると、画面に通知時の画像を表示する。
また、LEDが2回点滅する。WioTerminalの左ボタンを押すと画面がOFFになる。
なお、UDPポート番号は4416を使用している。
　こちらは、UDPで通知を受信したら液晶画面をONにして画像表示するだけなので、他の用途でも使えそう。

# ソース
.1台目 インターホンに取り付ける側
　notifier_DoorbellLightOn_wioterminalフォルダ

.2台目 通知を受けて画像を表示する側
　alerter_DoorbellLightOn_wioterminalフォルダ

# ビルド
tinygo version 0.39
go version 1.21 .. 1.25

## インターホンに取り付ける側
notifier_DoorbellLightOn_wioterminalフォルダにて
```
 go mod init main
 go mod tidy
 tinygo build --target=wioterminal main.go
```
 出来上がったuf2ファイルを1台目のWioTerminalのUSBストレージへコピー

## 通知を受けて画像を表示する側
　alerter_DoorbellLightOn_wioterminalフォルダにて
```
 go mod init main
 go mod tidy
 tinygo build --target=wioterminal main.go
```
 出来上がったuf2ファイルを2台目のWioTerminalのUSBストレージへコピー

## ビルド時の注意
"go mod init main"と"go mod tidy"は初回ビルド時のみ。
goバージョンが1.25.1の時に作成したgo.modとgo.sumファイルをコミットしているため、
別のバージョンを使用する場合は、go.modとgo.sumファイルを削除してから、"go mod init main"と"go mod tidy"を実施、その後でビルドする。
"go mod tidy"時にgithubに必要なパッケージを取りに行くため、ネット環境は必須。


## 使用している Gopher 画像
以下のURLを参照してください。
https://github.com/gotify/logo/blob/master/README.md
