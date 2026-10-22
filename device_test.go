package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func marshalConfig(t *testing.T, cfg ConfigData) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, cfg); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestConfigPacketLayout(t *testing.T) {
	var cfg ConfigData
	var cmd CommandData

	if got := binary.Size(cfg); got != 1041 {
		t.Fatalf("ConfigData size = %d, want 1041", got)
	}
	if got := binary.Size(cmd); got != 64 {
		t.Fatalf("CommandData size = %d, want 64", got)
	}

	cfg.PollingRateDivider = 0x15
	cfg.FilterFlags = 0x16
	cfg.Lod = 0x19
	cfg.AngleSnapping = true
	cfg.RippleControl = true
	cfg.MotionSync = true
	cfg.ActiveDpiStage = 0x1d
	cfg.CpiLevels = 0x1e
	cfg.StageColors[0] = 0x1f
	cfg.Cpis[0].XySplit = true
	cfg.Cpis[0].X = 0x3433
	cfg.Cpis[0].Y = 0x3635
	cfg.Handedness[0] = 0x47
	cfg.ButtonConfigs[0].Spdt = 0x4d
	cfg.ButtonConfigs[0].Mapping.Type = 0x4e
	cfg.ButtonConfigs[0].Mapping.Map[0] = 0x4f
	cfg.Reserved126 = 0x7e
	cfg.GlassMode = 0x7f
	cfg.AngleTuning = -1
	cfg.ForceMaxSensorFps = 0x81
	cfg.Reserved130 = 0x82
	cfg.Padding[0] = 0x83

	buf := marshalConfig(t, cfg)

	checks := []struct {
		name string
		got  byte
		want byte
	}{
		{"PollingRateDivider", buf[21], 0x15},
		{"FilterFlags", buf[22], 0x16},
		{"Lod", buf[25], 0x19},
		{"AngleSnapping", buf[26], 0x01},
		{"RippleControl", buf[27], 0x01},
		{"MotionSync", buf[28], 0x01},
		{"ActiveDpiStage", buf[29], 0x1d},
		{"CpiLevels", buf[30], 0x1e},
		{"StageColors", buf[31], 0x1f},
		{"CPI.XySplit", buf[51], 0x01},
		{"CPI.X low", buf[52], 0x33},
		{"CPI.X high", buf[53], 0x34},
		{"CPI.Y low", buf[54], 0x35},
		{"CPI.Y high", buf[55], 0x36},
		{"Handedness", buf[71], 0x47},
		{"ButtonConfig.Spdt", buf[77], 0x4d},
		{"ButtonConfig.Mapping.Type", buf[78], 0x4e},
		{"ButtonConfig.Mapping.Map", buf[79], 0x4f},
		{"Reserved126", buf[126], 0x7e},
		{"GlassMode", buf[127], 0x7f},
		{"AngleTuning", buf[128], 0xff},
		{"ForceMaxSensorFps", buf[129], 0x81},
		{"Reserved130", buf[130], 0x82},
		{"Padding", buf[131], 0x83},
	}

	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s byte = 0x%02x, want 0x%02x", check.name, check.got, check.want)
		}
	}
}

func TestSupportedDevices(t *testing.T) {
	for _, pid := range []uint16{XM28KProductID, XM28KV2ProductID} {
		if _, ok := findDevice(EndgameGearVendorID, pid); !ok {
			t.Errorf("PID 0x%04x not supported", pid)
		}
		if _, ok := findDevice(0x1234, pid); ok {
			t.Errorf("PID 0x%04x matched with wrong VID", pid)
		}
	}
}
