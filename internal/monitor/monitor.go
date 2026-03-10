package monitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Options 监控选项
type Options struct {
	Timestamp bool
	Hex       bool
	LogFile   string
	Filter    string
}

// Monitor 串口监控器
type Monitor struct {
	port    serial.Port
	options *Options
	logFile *os.File
}

// New 创建监控器
func New(port serial.Port, opts *Options) *Monitor {
	m := &Monitor{
		port:    port,
		options: opts,
	}

	if opts.LogFile != "" {
		f, err := os.OpenFile(opts.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			m.logFile = f
		}
	}

	return m
}

// ReadLoop 读取循环
func (m *Monitor) ReadLoop() {
	buf := make([]byte, 1024)
	lineBuf := make([]byte, 0, 4096)

	for {
		n, err := m.port.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n[串口已关闭]")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "读取错误: %v\n", err)
			continue
		}

		data := buf[:n]

		// 写入日志文件
		if m.logFile != nil {
			m.logFile.Write(data)
		}

		// 处理数据
		if m.options.Hex {
			m.printHex(data)
		} else {
			lineBuf = m.printText(data, lineBuf)
		}
	}
}

// WriteLoop 写入循环
func (m *Monitor) WriteLoop() {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		// 检查快捷键
		line = strings.TrimSuffix(line, "\n")
		if line == "\x12" { // Ctrl+R
			m.resetDevice()
			continue
		}
		if line == "\x0c" { // Ctrl+L
			fmt.Print("\033[2J\033[H") // 清屏
			continue
		}

		// 发送到串口
		m.port.Write([]byte(line + "\n"))
	}
}

// printText 打印文本
func (m *Monitor) printText(data []byte, lineBuf []byte) []byte {
	for _, b := range data {
		if b == '\n' {
			line := string(lineBuf)

			// 过滤
			if m.options.Filter != "" && !strings.Contains(line, m.options.Filter) {
				lineBuf = lineBuf[:0]
				continue
			}

			// 时间戳
			if m.options.Timestamp {
				fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05.000"), line)
			} else {
				fmt.Println(line)
			}

			lineBuf = lineBuf[:0]
		} else if b >= 32 || b == '\t' {
			lineBuf = append(lineBuf, b)
		}
	}
	return lineBuf
}

// printHex 打印十六进制
func (m *Monitor) printHex(data []byte) {
	for i, b := range data {
		if i%16 == 0 {
			if i > 0 {
				fmt.Println()
			}
			if m.options.Timestamp {
				fmt.Printf("[%s] ", time.Now().Format("15:04:05.000"))
			}
		}
		fmt.Printf("%02X ", b)
	}
	fmt.Println()
}

// resetDevice 复位设备
func (m *Monitor) resetDevice() {
	// DTR 触发复位
	m.port.SetDTR(true)
	time.Sleep(100 * time.Millisecond)
	m.port.SetDTR(false)
	fmt.Println("[设备已复位]")
}

// Close 关闭监控器
func (m *Monitor) Close() {
	if m.logFile != nil {
		m.logFile.Close()
	}
}
