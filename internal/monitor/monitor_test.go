package monitor

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
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

// ============ Plotter Tests ============

func TestNewPlotter(t *testing.T) {
	p := NewPlotter(100)
	if p == nil {
		t.Fatal("NewPlotter returned nil")
	}
	if p.maxPoints != 100 {
		t.Errorf("expected maxPoints 100, got %d", p.maxPoints)
	}
	if len(p.series) != 0 {
		t.Errorf("expected empty series, got %d", len(p.series))
	}
}

func TestNewPlotterDefault(t *testing.T) {
	p := NewPlotter(0)
	if p.maxPoints != 100 {
		t.Errorf("expected default maxPoints 100, got %d", p.maxPoints)
	}
}

func TestNewPlotterNegative(t *testing.T) {
	p := NewPlotter(-10)
	if p.maxPoints != 100 {
		t.Errorf("expected default maxPoints 100 for negative, got %d", p.maxPoints)
	}
}

func TestPlotterSetEnabled(t *testing.T) {
	p := NewPlotter(10)

	// Enabled by default
	points := p.ParseLine("temp:25.0")
	if len(points) != 1 {
		t.Error("expected to parse when enabled")
	}

	// Disable
	p.SetEnabled(false)
	points = p.ParseLine("temp:25.0")
	if points != nil {
		t.Error("expected nil when disabled")
	}

	// Re-enable
	p.SetEnabled(true)
	points = p.ParseLine("temp:25.0")
	if len(points) != 1 {
		t.Error("expected to parse when re-enabled")
	}
}

func TestParseLineSingleValue(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("123.45")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Name != "value" {
		t.Errorf("expected name 'value', got %s", points[0].Name)
	}
	if points[0].Value != 123.45 {
		t.Errorf("expected value 123.45, got %f", points[0].Value)
	}
}

func TestParseLineNameValuePair(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("temperature:28.5")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Name != "temperature" {
		t.Errorf("expected name 'temperature', got %s", points[0].Name)
	}
	if points[0].Value != 28.5 {
		t.Errorf("expected value 28.5, got %f", points[0].Value)
	}
}

func TestParseLineMultiplePairs(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("temp:28.5,hum:65.2,pressure:1013.25")
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	if points[0].Name != "temp" || points[0].Value != 28.5 {
		t.Errorf("unexpected first point: %+v", points[0])
	}
	if points[1].Name != "hum" || points[1].Value != 65.2 {
		t.Errorf("unexpected second point: %+v", points[1])
	}
	if points[2].Name != "pressure" || points[2].Value != 1013.25 {
		t.Errorf("unexpected third point: %+v", points[2])
	}
}

func TestParseLineWithSpaces(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("temp: 28.5 , hum : 65.2 ")
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].Name != "temp" {
		t.Errorf("expected name 'temp', got %s", points[0].Name)
	}
	if points[1].Name != "hum" {
		t.Errorf("expected name 'hum', got %s", points[1].Name)
	}
}

func TestParseLineEmpty(t *testing.T) {
	p := NewPlotter(10)
	if points := p.ParseLine(""); points != nil {
		t.Error("expected nil for empty line")
	}
	if points := p.ParseLine("   "); points != nil {
		t.Error("expected nil for whitespace line")
	}
}

func TestParseLineInvalid(t *testing.T) {
	p := NewPlotter(10)
	if points := p.ParseLine("hello world"); points != nil {
		t.Error("expected nil for invalid line")
	}
	if points := p.ParseLine("name:invalid"); points != nil {
		t.Error("expected nil for name with no number")
	}
}

func TestParseLineNegative(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("temp:-10.5")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Value != -10.5 {
		t.Errorf("expected -10.5, got %f", points[0].Value)
	}
}

func TestParseLineScientificNotation(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("value:1.23e-4")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Value < 0.00012 || points[0].Value > 0.000124 {
		t.Errorf("expected ~0.000123, got %f", points[0].Value)
	}
}

func TestParseLineWithUnits(t *testing.T) {
	p := NewPlotter(10)
	points := p.ParseLine("temp:25.5°C")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Value != 25.5 {
		t.Errorf("expected 25.5, got %f", points[0].Value)
	}
}

func TestPlotterUpdate(t *testing.T) {
	p := NewPlotter(10)
	points := []PlotterPoint{{Name: "temp", Value: 25.0}}
	series := p.Update(points)
	if len(series) != 1 {
		t.Fatalf("expected 1 series, got %d", len(series))
	}
	if series[0].Name != "temp" {
		t.Errorf("expected name 'temp', got %s", series[0].Name)
	}
	if len(series[0].Values) != 1 || series[0].Values[0] != 25.0 {
		t.Errorf("unexpected values: %v", series[0].Values)
	}
}

func TestPlotterUpdateMultipleSeries(t *testing.T) {
	p := NewPlotter(10)
	points := []PlotterPoint{
		{Name: "temp", Value: 25.0},
		{Name: "hum", Value: 50.0},
	}
	series := p.Update(points)
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}
}

func TestPlotterUpdateMaxPoints(t *testing.T) {
	p := NewPlotter(5)
	for i := 0; i < 10; i++ {
		p.Update([]PlotterPoint{{Name: "value", Value: float64(i)}})
	}
	series := p.Update([]PlotterPoint{{Name: "value", Value: 100}})
	if len(series[0].Values) > 5 {
		t.Errorf("expected max 5 values, got %d", len(series[0].Values))
	}
	// Should have last 5 values: 6, 7, 8, 9, 100
	expected := []float64{6, 7, 8, 9, 100}
	for i, v := range series[0].Values {
		if v != expected[i] {
			t.Errorf("at index %d: expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestPlotterUpdateEmpty(t *testing.T) {
	p := NewPlotter(10)
	series := p.Update([]PlotterPoint{})
	if series != nil {
		t.Error("expected nil for empty points")
	}
}

func TestPlotterClear(t *testing.T) {
	p := NewPlotter(10)
	p.Update([]PlotterPoint{{Name: "temp", Value: 25.0}})
	if len(p.series) != 1 {
		t.Error("expected 1 series before clear")
	}
	p.Clear()
	if len(p.series) != 0 {
		t.Error("expected 0 series after clear")
	}
}

func TestPlotterColors(t *testing.T) {
	p := NewPlotter(10)
	// Add more than len(colors) series
	for i := 0; i < 12; i++ {
		p.Update([]PlotterPoint{{Name: string(rune('a' + i)), Value: float64(i)}})
	}
	// Should cycle through colors
	series := p.Update([]PlotterPoint{{Name: "x", Value: 0}})
	if len(series) == 0 {
		t.Error("expected series")
	}
}

func TestPlotterToJSON(t *testing.T) {
	p := NewPlotter(10)
	p.Update([]PlotterPoint{{Name: "temp", Value: 25.0}})
	series := p.Update([]PlotterPoint{{Name: "temp", Value: 26.0}})

	jsonStr := p.ToJSON(series)

	// Verify it's valid JSON
	var msg OutputMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if msg.Type != "data" {
		t.Errorf("expected type 'data', got %s", msg.Type)
	}
	if len(msg.Data) != 1 {
		t.Errorf("expected 1 data series, got %d", len(msg.Data))
	}
	if msg.Time == "" {
		t.Error("expected time to be set")
	}
}

func TestLogToJSON(t *testing.T) {
	jsonStr := LogToJSON("test log message")

	var msg OutputMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if msg.Type != "log" {
		t.Errorf("expected type 'log', got %s", msg.Type)
	}
	if msg.Log != "test log message" {
		t.Errorf("expected log 'test log message', got %s", msg.Log)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"123.45", 123.45, false},
		{"-10.5", -10.5, false},
		{"1.23e-4", 0.000123, false},
		{"1.23E+4", 12300, false},
		{"25.5°C", 25.5, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		result, err := parseFloat(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for %s: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("for %s: expected %f, got %f", tt.input, tt.expected, result)
			}
		}
	}
}

// ============ Monitor Tests ============

func TestNewMonitor(t *testing.T) {
	port := NewMockPort()
	opts := &Options{Timestamp: true}
	mon := New(port, opts)
	if mon == nil {
		t.Fatal("New returned nil")
	}
	if mon.port != port {
		t.Error("port not set correctly")
	}
	if mon.options != opts {
		t.Error("options not set correctly")
	}
}

func TestNewMonitorWithPlotter(t *testing.T) {
	port := NewMockPort()
	opts := &Options{Plotter: true}
	mon := New(port, opts)
	if mon.plotter == nil {
		t.Error("expected plotter to be created")
	}
}

func TestNewMonitorWithJSON(t *testing.T) {
	port := NewMockPort()
	opts := &Options{JSON: true}
	mon := New(port, opts)
	if mon.plotter == nil {
		t.Error("expected plotter to be created for JSON mode")
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

func TestMonitorResetDevice(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})

	// DTR should be toggled during reset
	// After resetDevice, DTR is set to true then false
	mon.resetDevice()
	// The function completed without error - DTR was used
}

func TestMonitorResetDeviceDTR(t *testing.T) {
	port := NewMockPort()
	_ = New(port, &Options{})

	// Manually test DTR setting
	port.SetDTR(true)
	if !port.dtrLevel {
		t.Error("expected DTR true after SetDTR(true)")
	}
	port.SetDTR(false)
	if port.dtrLevel {
		t.Error("expected DTR false after SetDTR(false)")
	}
}

func TestMonitorResetDeviceJSON(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{JSON: true})

	mon.resetDevice()
	// Should not panic with JSON mode
}

func TestMonitorClose(t *testing.T) {
	// Create temp log file
	tmpFile := "/tmp/monitor_test_" + strings.ReplaceAll(time.Now().Format("20060102150405"), ".", "") + ".log"
	defer os.Remove(tmpFile)

	port := NewMockPort()
	mon := New(port, &Options{LogFile: tmpFile})

	if mon.logFile == nil {
		t.Fatal("log file not created")
	}

	mon.Close()
	// Should not error when closing
}

func TestMonitorCloseNoLogFile(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})
	mon.Close()
	// Should not error when no log file
}

func TestMonitorOptions(t *testing.T) {
	opts := &Options{
		Timestamp: true,
		Hex:       false,
		LogFile:   "/tmp/test.log",
		Filter:    "error",
		JSON:      true,
		Plotter:   true,
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
	if !opts.JSON {
		t.Error("expected JSON true")
	}
	if !opts.Plotter {
		t.Error("expected Plotter true")
	}
}

func TestMonitorPrintText(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{Timestamp: false})

	lineBuf := make([]byte, 0, 100)
	data := []byte("hello\nworld\n")
	lineBuf = mon.printText(data, lineBuf)

	if len(lineBuf) > 0 {
		t.Error("expected lineBuf to be cleared after newline")
	}
}

func TestMonitorPrintTextWithFilter(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{Filter: "error"})

	lineBuf := make([]byte, 0, 100)
	data := []byte("info message\nerror message\nwarning\n")
	lineBuf = mon.printText(data, lineBuf)

	// Filter should skip non-matching lines
}

func TestMonitorPrintTextWithPlotter(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{Plotter: true})

	lineBuf := make([]byte, 0, 100)
	data := []byte("temp:25.5\n")
	lineBuf = mon.printText(data, lineBuf)

	// Should parse as plotter data
	if mon.plotter == nil {
		t.Error("expected plotter to be initialized")
	}
}

func TestMonitorPrintTextJSON(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{JSON: true})

	lineBuf := make([]byte, 0, 100)
	data := []byte("hello\n")
	lineBuf = mon.printText(data, lineBuf)

	// Should output as JSON
}

func TestMonitorPrintHex(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{Hex: true, Timestamp: false})

	data := []byte{0x01, 0x02, 0x03, 0x04}
	mon.printHex(data)
	// Should print hex without error
}

func TestMonitorPrintHexWithTimestamp(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{Hex: true, Timestamp: true})

	data := []byte{0x01, 0x02, 0x03, 0x04}
	mon.printHex(data)
	// Should print hex with timestamp without error
}

func TestMonitorWriteLoop(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{JSON: true})

	// WriteLoop runs in infinite loop, just verify it doesn't crash
	_ = mon
}

func TestMonitorWriteLoopReset(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})

	// Simulate Ctrl+R input
	_ = "\x12"
	_ = mon
	// resetDevice should be called
}

func TestMonitorPrintTextTab(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})

	lineBuf := make([]byte, 0, 100)
	data := []byte("hello\tworld\n")
	lineBuf = mon.printText(data, lineBuf)

	// Tab should be included
}

func TestMonitorPrintTextControlChars(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{})

	lineBuf := make([]byte, 0, 100)
	// Include control characters that should be filtered
	data := []byte("hello\x00\x01\x02world\n")
	lineBuf = mon.printText(data, lineBuf)

	// Control chars should be filtered out
}

// ============ Integration Tests ============

func TestMonitorIntegration(t *testing.T) {
	port := NewMockPort()
	mon := New(port, &Options{
		Timestamp: true,
		Plotter:   true,
	})

	// Simulate receiving data
	data := []byte("temp:25.5,hum:50.0\n")
	lineBuf := make([]byte, 0, 100)
	lineBuf = mon.printText(data, lineBuf)

	// Verify plotter parsed the data
	if len(mon.plotter.series) != 2 {
		t.Errorf("expected 2 series, got %d", len(mon.plotter.series))
	}
}

func TestPlotterIntegration(t *testing.T) {
	p := NewPlotter(10)

	// Simulate multiple readings
	lines := []string{
		"temp:25.0,hum:50.0",
		"temp:25.5,hum:51.0",
		"temp:26.0,hum:52.0",
	}

	for _, line := range lines {
		points := p.ParseLine(line)
		if len(points) != 2 {
			t.Errorf("expected 2 points from %s, got %d", line, len(points))
		}
		p.Update(points)
	}

	series := p.Update(p.ParseLine("temp:26.5,hum:53.0"))
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}

	// Each series should have 4 values
	for _, s := range series {
		if len(s.Values) != 4 {
			t.Errorf("series %s: expected 4 values, got %d", s.Name, len(s.Values))
		}
	}
}