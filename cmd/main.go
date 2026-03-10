package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.bug.st/serial"

	"github.com/zyrthi-io/zyrthi-monitor/internal/config"
	"github.com/zyrthi-io/zyrthi-monitor/internal/monitor"
)

var (
	flagConfig    = flag.String("config", "zyrthi.yaml", "配置文件路径")
	flagPort      = flag.String("port", "", "串口设备")
	flagBaud      = flag.Int("baud", 0, "波特率")
	flagTimestamp = flag.Bool("timestamp", false, "显示时间戳")
	flagHex       = flag.Bool("hex", false, "十六进制显示")
	flagLog       = flag.String("log", "", "日志保存文件")
	flagFilter    = flag.String("filter", "", "过滤关键字")
)

func main() {
	flag.Parse()

	// 读取配置
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 无法读取配置文件: %v\n", err)
		cfg = &config.Config{Monitor: config.MonitorConfig{Baud: 115200}}
	}

	// 参数覆盖配置
	baud := *flagBaud
	if baud == 0 {
		baud = cfg.Monitor.Baud
		if baud == 0 {
			baud = 115200
		}
	}

	port := *flagPort
	if port == "" {
		port = cfg.Monitor.Port
		if port == "" {
			// 自动检测
			ports, err := serial.GetPortsList()
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: 无法获取串口列表: %v\n", err)
				os.Exit(1)
			}
			if len(ports) == 0 {
				fmt.Fprintln(os.Stderr, "错误: 未找到串口设备")
				os.Exit(1)
			}
			port = ports[0]
			fmt.Printf("自动选择串口: %s\n", port)
		}
	}

	// 打开串口
	serialPort, err := serial.Open(port, &serial.Mode{
		BaudRate: baud,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法打开串口 %s: %v\n", port, err)
		os.Exit(1)
	}
	defer serialPort.Close()

	fmt.Printf("已连接 %s @ %d baud\n", port, baud)
	fmt.Println("按 Ctrl+C 退出, Ctrl+R 复位设备")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动读写
	mon := monitor.New(serialPort, &monitor.Options{
		Timestamp: *flagTimestamp || cfg.Monitor.Timestamp,
		Hex:       *flagHex,
		LogFile:   *flagLog,
		Filter:    *flagFilter,
	})

	go mon.ReadLoop()
	go mon.WriteLoop()

	// 等待退出
	<-sigChan
	fmt.Println("\n退出")
}
