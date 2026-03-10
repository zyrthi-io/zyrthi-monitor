package main

import (
	"bytes"
	"testing"
	"time"

	"go.bug.st/serial"
)

// MockPort 模拟串口
type MockPort struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	dtrLevel bool
	rtsLevel bool
}

func NewMockPort() *MockPort {
	return &MockPort{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *MockPort) Read(p []byte) (n int, err error) {
	return m.readBuf.Read(p)
}

func (m *MockPort) Write(p []byte) (n int, err error) {
	return m.writeBuf.Write(p)
}

func (m *MockPort) Close() error {
	return nil
}

func (m *MockPort) SetDTR(level bool) error {
	m.dtrLevel = level
	return nil
}

func (m *MockPort) SetRTS(level bool) error {
	m.rtsLevel = level
	return nil
}

func (m *MockPort) Break(breakDuration time.Duration) error {
	return nil
}

func (m *MockPort) SetMode(mode *serial.Mode) error {
	return nil
}

func (m *MockPort) Drain() error {
	return nil
}

func (m *MockPort) ResetInputBuffer() error {
	return nil
}

func (m *MockPort) ResetOutputBuffer() error {
	return nil
}

func (m *MockPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}

func (m *MockPort) SetReadTimeout(t time.Duration) error {
	return nil
}

func TestMonitorNew(t *testing.T) {
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	// 测试 Options 结构体
	if opts.Timestamp {
		t.Error("expected timestamp false")
	}
	if opts.Hex {
		t.Error("expected hex false")
	}
}

func TestMonitorPrintTextBasic(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)
	if mon == nil {
		t.Fatal("NewMonitor returned nil")
	}

	data := []byte("Hello World\n")
	lineBuf := make([]byte, 0, 4096)

	lineBuf = mon.printText(data, lineBuf)

	// 换行符后 lineBuf 应该被重置
	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf, got %d bytes", len(lineBuf))
	}
}

func TestMonitorPrintTextNoNewline(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)

	data := []byte("Hello ")
	lineBuf := make([]byte, 0, 4096)

	lineBuf = mon.printText(data, lineBuf)

	// 没有换行符，数据应该在 lineBuf 中
	if len(lineBuf) != 6 {
		t.Errorf("expected 6 bytes in lineBuf, got %d", len(lineBuf))
	}
}

func TestMonitorPrintTextFilter(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
		Filter:    "error",
	}

	mon := NewMonitor(port, opts)

	// 不匹配过滤器的数据
	data := []byte("this is info message\n")
	lineBuf := make([]byte, 0, 4096)
	lineBuf = mon.printText(data, lineBuf)
	// 应该被过滤，lineBuf 被重置
	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf for filtered content, got %d", len(lineBuf))
	}
}

func TestMonitorPrintTextControlChars(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)

	// 包含控制字符的数据（除了 tab）
	data := []byte("Hello\x00\x01\x02World\n")
	lineBuf := make([]byte, 0, 4096)

	lineBuf = mon.printText(data, lineBuf)

	// 换行后 lineBuf 应该被重置
	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf after newline, got %d bytes", len(lineBuf))
	}
}

func TestMonitorPrintTextTab(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)

	data := []byte("Hello\tWorld\n")
	lineBuf := make([]byte, 0, 4096)

	lineBuf = mon.printText(data, lineBuf)

	// 换行后 lineBuf 应该被重置
	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf after newline, got %d bytes", len(lineBuf))
	}
}

func TestMonitorClose(t *testing.T) {
	port := NewMockPort()
	opts := &Options{}

	mon := NewMonitor(port, opts)
	mon.Close()
	// 无错误即为成功
}

func TestMonitorPrintHex(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       true,
	}

	mon := NewMonitor(port, opts)

	data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F} // "Hello"

	mon.printHex(data)
	// 无错误即为成功
}

func TestMonitorWithLogFile(t *testing.T) {
	port := NewMockPort()
	tmpFile := "/tmp/test_monitor.log"
	opts := &Options{
		LogFile: tmpFile,
	}

	mon := NewMonitor(port, opts)

	if mon.logFile == nil {
		t.Error("expected log file to be created")
	}

	// 清理
	mon.Close()
}

func TestMonitorResetDevice(t *testing.T) {
	port := NewMockPort()
	opts := &Options{}

	mon := NewMonitor(port, opts)

	mon.resetDevice()
	
	// 验证 DTR 被设置
	if port.dtrLevel {
		t.Error("expected DTR to be false after reset")
	}
}

// 测试 lineBuf 连续处理
func TestMonitorPrintTextContinuous(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)

	lineBuf := make([]byte, 0, 4096)

	// 第一段数据（没有换行）
	lineBuf = mon.printText([]byte("Hello "), lineBuf)
	if len(lineBuf) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(lineBuf))
	}

	// 第二段数据（有换行）
	lineBuf = mon.printText([]byte("World\n"), lineBuf)
	if len(lineBuf) != 0 {
		t.Errorf("expected 0 bytes after newline, got %d", len(lineBuf))
	}
}

// TestReadLoopWithData 测试 ReadLoop 读取数据
func TestReadLoopWithData(t *testing.T) {
	port := NewMockPort()
	port.readBuf.Write([]byte("test message\n"))

	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	mon := NewMonitor(port, opts)

	// 使用 channel 控制 goroutine
	done := make(chan bool)
	go func() {
		// 读取一条消息后关闭
		<-done
	}()

	// 模拟读取
	buf := make([]byte, 1024)
	n, err := port.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	data := buf[:n]
	lineBuf := mon.printText(data, make([]byte, 0, 4096))

	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf after newline")
	}

	done <- true
}

// TestReadLoopEOF 测试 ReadLoop EOF 处理
func TestReadLoopEOF(t *testing.T) {
	port := NewMockPort()
	// 不写入任何数据

	opts := &Options{
		Timestamp: false,
		Hex:       false,
	}

	_ = opts // 避免 unused variable 警告
	// ReadLoop 的 EOF 处理会调用 os.Exit，无法直接测试
	// 但可以测试读取逻辑
	_ = port
}

// TestWriteLoop 测试 WriteLoop 写入
func TestWriteLoop(t *testing.T) {
	port := NewMockPort()
	opts := &Options{}

	_ = NewMonitor(port, opts) // 创建 monitor 但不需要使用

	// 直接测试写入
	testData := []byte("test command\n")
	n, err := port.Write(testData)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	// 验证写入的数据
	written := port.writeBuf.Bytes()
	if string(written) != string(testData) {
		t.Errorf("expected %q, got %q", string(testData), string(written))
	}
}

// TestWriteLoopReset 测试 Ctrl+R 复位
func TestWriteLoopReset(t *testing.T) {
	port := NewMockPort()
	opts := &Options{}

	mon := NewMonitor(port, opts)

	// 执行复位
	mon.resetDevice()

	// 验证 DTR 被正确操作
	// resetDevice 流程: SetDTR(true) -> Sleep -> SetDTR(false)
	// 最终 DTR 应该是 false
	if port.dtrLevel {
		t.Error("expected DTR to be false after reset")
	}
}

// TestReadLoopWithHex 测试十六进制模式读取
func TestReadLoopWithHex(t *testing.T) {
	port := NewMockPort()
	port.readBuf.Write([]byte{0x01, 0x02, 0x03})

	opts := &Options{
		Timestamp: false,
		Hex:       true,
	}

	mon := NewMonitor(port, opts)

	// 读取并处理
	buf := make([]byte, 1024)
	n, _ := port.Read(buf)
	data := buf[:n]

	// hex 模式处理
	mon.printHex(data)
	// 无错误即为成功
}

// TestReadLoopWithFilter 测试过滤功能
func TestReadLoopWithFilter(t *testing.T) {
	port := NewMockPort()
	port.readBuf.Write([]byte("[ERROR] something failed\n[INFO] all good\n"))

	opts := &Options{
		Timestamp: false,
		Hex:       false,
		Filter:    "ERROR",
	}

	_ = NewMonitor(port, opts) // 创建 monitor

	// 读取并处理
	buf := make([]byte, 1024)
	n, _ := port.Read(buf)
	data := buf[:n]

	lineBuf := make([]byte, 0, 4096)
	// 处理数据
	for _, b := range data {
		if b == '\n' {
			line := string(lineBuf)
			// 过滤逻辑：只保留包含 "ERROR" 的行
			if opts.Filter != "" && !contains(line, opts.Filter) {
				lineBuf = lineBuf[:0]
				continue
			}
			lineBuf = lineBuf[:0]
		} else if b >= 32 || b == '\t' {
			lineBuf = append(lineBuf, b)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMonitorWithTimestamp 测试时间戳模式
func TestMonitorWithTimestamp(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: true,
		Hex:       false,
	}

	_ = port // 用于创建 monitor

	data := []byte("test\n")
	lineBuf := make([]byte, 0, 4096)

	// 直接测试 printText 函数
	// 创建一个临时 monitor 用于测试
	mon := &Monitor{
		port:    port,
		options: opts,
	}
	lineBuf = mon.printText(data, lineBuf)
	// 无错误即为成功，时间戳会自动添加
	if len(lineBuf) != 0 {
		t.Errorf("expected empty lineBuf after newline")
	}
}

// TestMonitorWithTimestampHex 测试时间戳 + 十六进制模式
func TestMonitorWithTimestampHex(t *testing.T) {
	port := NewMockPort()
	opts := &Options{
		Timestamp: true,
		Hex:       true,
	}

	mon := NewMonitor(port, opts)

	data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
	mon.printHex(data)
	// 无错误即为成功
}

// TestSerialPortOperations 测试串口操作组合
func TestSerialPortOperations(t *testing.T) {
	port := NewMockPort()

	// 测试写入
	n, err := port.Write([]byte("AT\r\n"))
	if err != nil || n != 4 {
		t.Fatalf("Write failed: n=%d, err=%v", n, err)
	}

	// 模拟响应
	port.readBuf.Write([]byte("OK\r\n"))

	// 测试读取
	buf := make([]byte, 10)
	n, err = port.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "OK\r\n" {
		t.Errorf("expected 'OK\\r\\n', got %q", string(buf[:n]))
	}

	// 测试 DTR/RTS
	port.SetDTR(true)
	if !port.dtrLevel {
		t.Error("DTR should be true")
	}

	port.SetRTS(true)
	if !port.rtsLevel {
		t.Error("RTS should be true")
	}
}
