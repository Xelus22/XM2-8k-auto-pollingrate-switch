package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

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
	Pad2               [1]byte         //29
	CpiLevels          uint8           //30
	Pad3               [20]byte        //31
	Cpis               [4]CPI          //51
	Pad4               [6]byte         //71
	ButtonConfigs      [7]ButtonConfig //77
	CustomFlags        uint16          //154
	Pad5               [913]byte       //992
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

var config ConfigData

var bChanged bool = false

func getDeviceInfo() (string, error) {

	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		// Append the device to the list by copying the struct
		if info.VendorID == 0x3367 && info.ProductID == 0x1966 && info.Usage == 0x0002 && info.UsagePage == 0xFF01 {
			mousePath = info.Path
		}
		return nil
	})

	fmt.Println("Mouse Path: ", mousePath)
	device, err := hid.OpenPath(mousePath)

	if err != nil {
		panic(err)
	}

	defer device.Close()

	// write a feature report to the device to get the fw version
	report := CommandData{
		Op: OpCodeGetFwVersion,
	}

	// convert report to byte array
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, report)
	if err != nil {
		panic(err)
	}

	// print buf length
	fmt.Println("Buf Length: ", len(buf.Bytes()))

	var num int
	num, err = device.SendFeatureReport(buf.Bytes())
	if err != nil {
		// print error
		// print num
		// print buf length
		fmt.Println("Num: ", num)
		panic(err)
	}

	if num != len(buf.Bytes()) {
		panic("Failed to send feature report")
	}

	// Read a feature report from the device
	buf.Reset()
	config = ConfigData{
		Op: OpCodeNone,
	}
	err = binary.Write(&buf, binary.LittleEndian, config)

	num, err = device.GetFeatureReport(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// fmt.Println(buf)

	err = binary.Read(&buf, binary.LittleEndian, &config)

	// version
	var version string
	version = fmt.Sprintf("%x.%x", config.Pad0[16], config.Pad0[15])
	fmt.Println(version)
	return version, nil
}

func getConfig() {
	if mousePath == "" {
		getDeviceInfo()
	}

	// open device
	device, err := hid.OpenPath(mousePath)

	if err != nil {
		panic(err)
	}

	defer device.Close()

	// write loadConfig
	report := CommandData{
		Op: OpCodeLoadConfig,
	}

	// convert report to byte array
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, report)
	if err != nil {
		panic(err)
	}

	device.SendFeatureReport(buf.Bytes())

	// Read a feature report from the device
	buf.Reset()

	err = binary.Write(&buf, binary.LittleEndian, config)
	var num int
	num, err = device.GetFeatureReport(buf.Bytes())
	if err != nil {
		fmt.Println(num)
		panic(err)
	}

	err = binary.Read(&buf, binary.LittleEndian, &config)

	// convert config to bytes
	err = binary.Write(&buf, binary.LittleEndian, config)
	if err != nil {
		panic(err)
	}

	// print length of buf
	fmt.Println("Config Buf Length: ", len(buf.Bytes()))
	// print out some buf bytes
	fmt.Println(buf.Bytes())

	fmt.Println("Read Config")

	config.Pad2[0] = 0x02
	copy(config.Pad3[0:], []byte{0xff, 0xff, 0x0, 0x1, 0x1, 0x0, 0x0, 0xff, 0x1, 0x2, 0xff, 0x0, 0x0, 0x1, 0x3, 0x0, 0xff, 0x0, 0x1, 0x4})
	copy(config.Pad4[0:], []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x0})

	// read each button config
	for i := 0; i < 7; i++ {
		fmt.Println("Button Config ", i)
		fmt.Println("SPDT: ", config.ButtonConfigs[i].Spdt)
		fmt.Println("Mapping Type: ", config.ButtonConfigs[i].Mapping.Type)
		fmt.Println("Mapping: ", config.ButtonConfigs[i].Mapping.Map)
	}
}

func set8k() {
	if config.PollingRateDivider == 1 {
		return
	}
	config.PollingRateDivider = 1
	fmt.Println("Config Set to 8k")
	bChanged = true
}

func set1k() {
	if config.PollingRateDivider == 8 {
		return
	}
	config.PollingRateDivider = 8
	fmt.Println("Config Set to 1k")
	bChanged = true
}

func setConfig() {
	if !bChanged {
		return
	}

	if mousePath == "" {
		getDeviceInfo()
	}

	// open device
	device, err := hid.OpenPath(mousePath)

	if err != nil {
		panic(err)
	}

	defer device.Close()

	// write loadConfig
	config.Op = OpCodeStoreConfig

	// convert report to byte array
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, config)
	if err != nil {
		panic(err)
	}

	num, err := device.SendFeatureReport(buf.Bytes())
	if err != nil {
		panic(err)
	}
	fmt.Println("num sent:", num)

	bChanged = false

	fmt.Println("Config Set")
}
