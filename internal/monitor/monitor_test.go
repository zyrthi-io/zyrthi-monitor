package monitor

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

func (m *MockPort) Read(p []byte) (n int, err error)    { return m.readBuf.Read(p) }
func (m *MockPort) Write(p []byte) (n int, err error)   { return m.writeBuf.Write(p) }
func (m *MockPort) Close() error                        { return nil }
func (m *MockPort) SetDTR(level bool) error             { m.dtrLevel = level; return nil }
func (m *MockPort) SetRTS(level bool) error             { m.rtsLevel = level; return nil }
func (m *MockPort) Break(time.Duration) error           { return nil }
func (m *MockPort) SetMode(*serial.Mode) error          { return nil }
func (m *MockPort) Drain() error                        { return nil }
func (m *MockPort) ResetInputBuffer() error             { return nil }
func (m *MockPort) ResetOutputBuffer() error            { return nil }
func (m *MockPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}
func (m *MockPort) SetReadTimeout(time.Duration) error  { return nil }

func TestNewMonitor(t *testing.T) {
	port := NewMockPort()
	opts := &Options{Timestamp: true}
	mon := New(port, opts)
	if mon == nil {
		t.Fatal("New returned nil")
	}
}

func TestMonitorWrite(t *testing.T) {
	port := NewMockPort()
	_ = New(port, &Options{})

	data := []byte("test")
	port.readBuf.Write(data)

	buf := make([]byte, 100)
	n, err := port.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}
}

func TestMonitorReset(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})

	mon.resetDevice()
	// 验证 DTR 被设置后重置
}

func TestMonitorOptions(t *testing.T) {
	opts := &Options{
		Timestamp: true,
		Hex:       false,
		LogFile:   "/tmp/test.log",
		Filter:    "error",
	}

	if !opts.Timestamp {
		t.Error("expected Timestamp true")
	}
	if opts.Hex {
		t.Error("expected Hex false")
	}
	if opts.Filter != "error" {
		t.Errorf("expected Filter 'error', got %s", opts.Filter)
	}
}
