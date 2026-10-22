package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/sstallion/go-hid"
)

type OpCode uint16
type SPDTMode uint8
type MappingType int8

const (
	OpCodeNone         OpCode = 0x00a1
	OpCodeStoreConfig  OpCode = 0x11a0
	OpCodeLoadConfig   OpCode = 0x12a1
	OpCodeFactoryReset OpCode = 0x13a1
	OpCodeGetFwVersion OpCode = 0x02a1
)

const (
	EndgameGearVendorID uint16 = 0x3367
	XM28KProductID      uint16 = 0x1966
	XM28KV2ProductID    uint16 = 0x1980
	configReportID      byte   = 0xa0
	commandReportID     byte   = 0xa1
)

type deviceProfile struct {
	PID  uint16
	Name string
}

var supportedDevices = []deviceProfile{
	{PID: XM28KProductID, Name: "XM2 8k"},
	{PID: XM28KV2ProductID, Name: "XM2 8k v2"},
}

type CPI struct {
	XySplit bool
	X       uint16
	Y       uint16
}
type Mapping struct {
	Type MappingType
	Map  [5]byte
} // 5 bytes

type ButtonConfig struct {
	Spdt    uint8   // 1 byte
	Mapping Mapping // 10 bytes
}

type ConfigData struct {
	Op                 OpCode          //0
	Pad0               [19]byte        //2
	PollingRateDivider uint8           //21
	FilterFlags        uint8           //22
	Pad1               [2]byte         //23
	Lod                uint8           //25
	AngleSnapping      bool            //26
	RippleControl      bool            //27
	MotionSync         bool            //28
	ActiveDpiStage     uint8           //29
	CpiLevels          uint8           //30
	StageColors        [20]byte        //31
	Cpis               [4]CPI          //51
	Handedness         [6]byte         //71
	ButtonConfigs      [7]ButtonConfig //77
	Reserved126        uint8           //126
	GlassMode          uint8           //127
	AngleTuning        int8            //128
	ForceMaxSensorFps  uint8           //129
	Reserved130        uint8           //130
	Padding            [910]byte       //131
}

const (
	SPDTModeOff   uint8 = 0
	SPDTModeSafe  uint8 = 0xf0
	SPDTModeSpeed uint8 = 0xf1
)

const (
	MappingTypeMouse    MappingType = 0
	MappingTypeScroll   MappingType = 1
	MappingTypeKeyboard MappingType = 2
	MappingTypeCpiLoop  MappingType = 9
	MappingTypeCpi      MappingType = 12
	MappingTypeMedia    MappingType = 32
	MappingTypeDisable  MappingType = -1
)

type MouseKeys uint8

const (
	MouseKeysLeft    MouseKeys = 1
	MouseKeysRight   MouseKeys = 2
	MouseKeysMiddle  MouseKeys = 4
	MouseKeysBack    MouseKeys = 8
	MouseKeysForward MouseKeys = 16
)

type ScrollWheel int8

const (
	ScrollWheelUp   ScrollWheel = 1
	ScrollWheelDown ScrollWheel = -1
)

type MediaKeys uint8

const (
	MediaKeysPlayPause  MediaKeys = 0xcd
	MediaKeysNext       MediaKeys = 0xb5
	MediaKeysPrevious   MediaKeys = 0xb6
	MediaKeysMute       MediaKeys = 0xe2
	MediaKeysVolumeUp   MediaKeys = 0xe9
	MediaKeysVolumeDown MediaKeys = 0xea
	MediaKeysBrowser    MediaKeys = 0x96
	MediaKeysExplorer   MediaKeys = 0x94
)

type CommandData struct {
	Op  OpCode
	Pad [62]byte
}

var mousePath string
var deviceName string

var config ConfigData

var bChanged bool = false

func findDevice(vid uint16, pid uint16) (deviceProfile, bool) {
	if vid != EndgameGearVendorID {
		return deviceProfile{}, false
	}
	for _, device := range supportedDevices {
		if device.PID == pid {
			return device, true
		}
	}
	return deviceProfile{}, false
}

func getDeviceInfo() (string, error) {
	mousePath = ""
	deviceName = ""

	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if info.Usage == 0x0002 && info.UsagePage == 0xFF01 {
			if device, ok := findDevice(info.VendorID, info.ProductID); ok {
				mousePath = info.Path
				deviceName = device.Name
			}
		}
		return nil
	})

	debugPrintln("Mouse Path: ", mousePath)
	if mousePath == "" {
		return "", fmt.Errorf("no supported Endgame Gear device found")
	}
	debugPrintln("Device: ", deviceName)

	return getFirmwareVersion()
}

func getFirmwareVersion() (string, error) {
	device, err := hid.OpenPath(mousePath)
	if err != nil {
		return "", err
	}
	defer device.Close()

	report := CommandData{
		Op: OpCodeGetFwVersion,
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, report); err != nil {
		return "", err
	}

	num, err := device.SendFeatureReport(buf.Bytes())
	if err != nil {
		return "", err
	}
	if num != len(buf.Bytes()) {
		return "", fmt.Errorf("sent %d of %d firmware command bytes", num, len(buf.Bytes()))
	}

	time.Sleep(50 * time.Millisecond)

	buf.Reset()
	response := CommandData{
		Op: OpCodeNone,
	}
	if err := binary.Write(&buf, binary.LittleEndian, response); err != nil {
		return "", err
	}

	num, err = device.GetFeatureReport(buf.Bytes())
	if err != nil {
		return "", err
	}
	if num < 2 {
		return "", fmt.Errorf("short firmware response: %d bytes", num)
	}

	version := formatFirmwareVersion(buf.Bytes())
	debugPrintln("Firmware: ", version)
	return version, nil
}

func formatFirmwareVersion(buf []byte) string {
	for _, offset := range []int{17, 16, 18, 2, 4, 6} {
		if offset+1 >= len(buf) {
			continue
		}
		patch := buf[offset]
		major := buf[offset+1]
		if major >= 1 && major <= 9 && patch <= 99 {
			return fmt.Sprintf("V%d.%02d", major, patch)
		}
	}
	for _, offset := range []int{17, 16, 18, 2, 4, 6} {
		if offset+1 >= len(buf) {
			continue
		}
		value := int(buf[offset]) | int(buf[offset+1])<<8
		if value >= 50 && value <= 5000 {
			return fmt.Sprintf("V%.2f", float64(value)/100)
		}
	}
	return "unknown"
}

func sendCommand(device *hid.Device, op OpCode) error {
	command := CommandData{
		Op: op,
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, command); err != nil {
		return err
	}

	num, err := device.SendFeatureReport(buf.Bytes())
	if err != nil {
		return err
	}
	if num != len(buf.Bytes()) {
		return fmt.Errorf("sent %d of %d command bytes", num, len(buf.Bytes()))
	}
	return nil
}

func readConfigReport(device *hid.Device, reportID byte) (ConfigData, error) {
	var cfg ConfigData
	cfg.Op = OpCode(reportID)

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, cfg); err != nil {
		return cfg, err
	}

	num, err := device.GetFeatureReport(buf.Bytes())
	if err != nil {
		return cfg, err
	}
	if num < 31 {
		return cfg, fmt.Errorf("short config response: %d bytes", num)
	}

	if err := binary.Read(&buf, binary.LittleEndian, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func configLooksValid(cfg *ConfigData) bool {
	return uint8(cfg.Op>>8) == 0x01 &&
		cfg.CpiLevels >= 1 && cfg.CpiLevels <= 4 &&
		cfg.ActiveDpiStage <= 3 &&
		cfg.PollingRateDivider >= 1 && cfg.PollingRateDivider <= 64 &&
		cfg.Cpis[0].X >= 10 && cfg.Cpis[0].X <= 30000
}

func getConfig() error {
	if mousePath == "" {
		if _, err := getDeviceInfo(); err != nil {
			return err
		}
	}

	device, err := hid.OpenPath(mousePath)
	if err != nil {
		return err
	}
	defer device.Close()

	var lastErr error
	for round := 0; round < 3; round++ {
		if err := sendCommand(device, OpCodeLoadConfig); err != nil {
			lastErr = err
			continue
		}
		time.Sleep(80 * time.Millisecond)

		delay := 60 * time.Millisecond
		for poll := 0; poll < 6; poll++ {
			for _, reportID := range []byte{configReportID, commandReportID} {
				cfg, err := readConfigReport(device, reportID)
				if err == nil && configLooksValid(&cfg) {
					config = cfg
					debugPrintln("Read Config")
					for i := 0; i < 7; i++ {
						debugPrintln("Button Config ", i)
						debugPrintln("SPDT: ", config.ButtonConfigs[i].Spdt)
						debugPrintln("Mapping Type: ", config.ButtonConfigs[i].Mapping.Type)
						debugPrintln("Mapping: ", config.ButtonConfigs[i].Mapping.Map)
					}
					return nil
				}
				if err != nil {
					lastErr = err
				} else {
					lastErr = fmt.Errorf("invalid config frame from report 0x%02x", reportID)
				}
			}

			time.Sleep(delay)
			if delay < 400*time.Millisecond {
				delay *= 2
				if delay > 400*time.Millisecond {
					delay = 400 * time.Millisecond
				}
			}
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no config response")
	}
	return fmt.Errorf("failed to read config: %w", lastErr)
}

func set8k() {
	if config.PollingRateDivider == 1 {
		return
	}
	config.PollingRateDivider = 1
	SetCurrentPollingRate(POLLING_RATE_8K)
	debugPrintln("Config Set to 8k")
	bChanged = true
}

func set4k() {
	if config.PollingRateDivider == 2 {
		return
	}
	config.PollingRateDivider = 2
	SetCurrentPollingRate(POLLING_RATE_4K)
	debugPrintln("Config Set to 4k")
	bChanged = true
}

func set1k() {
	if config.PollingRateDivider == 8 {
		return
	}
	config.PollingRateDivider = 8
	SetCurrentPollingRate(POLLING_RATE_1K)
	debugPrintln("Config Set to 1k")
	bChanged = true
}

func setPollingRate(rate uint8) {
	switch rate {
	case 8:
		set8k()
	case 4:
		set4k()
	case 1:
		set1k()
	}
}

func setConfig() {
	if !bChanged {
		return
	}

	if mousePath == "" {
		if _, err := getDeviceInfo(); err != nil {
			debugPrintln("Error: ", err)
			return
		}
	}

	device, err := hid.OpenPath(mousePath)
	if err != nil {
		debugPrintln("Error: ", err)
		return
	}
	defer device.Close()

	config.Op = OpCodeStoreConfig

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, config); err != nil {
		debugPrintln("Error: ", err)
		return
	}

	num, err := device.SendFeatureReport(buf.Bytes())
	if err != nil {
		debugPrintln("Error: ", err)
		return
	}
	debugPrintln("num sent:", num)

	bChanged = false

	debugPrintln("Config Set")
}
