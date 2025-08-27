package utils

import (
	"crypto/md5"
	"encoding/hex"
	"unsafe"

	"github.com/alexbrainman/printer"
	"golang.org/x/sys/windows"
)

var printerMap = map[string]string{} // md5 -> name

type PrinterInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	PaperSizes []string `json:"paper_sizes"`
}

func HashPrinterName(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func GetSupportedPaperSizes(printerName string) []string {
	user32 := windows.NewLazySystemDLL("winspool.drv")
	proc := user32.NewProc("DeviceCapabilitiesW")

	const DC_PAPERNAMES = 16
	const MAX_PAPER_NAME = 64

	ret, _, _ := proc.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(printerName))),
		0,
		uintptr(DC_PAPERNAMES),
		0,
		0,
	)

	count := int(ret)
	if count <= 0 {
		return []string{}
	}

	buffer := make([]uint16, count*MAX_PAPER_NAME)

	proc.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(printerName))),
		0,
		uintptr(DC_PAPERNAMES),
		uintptr(unsafe.Pointer(&buffer[0])),
		0,
	)

	var sizes []string
	for i := 0; i < count; i++ {
		start := i * MAX_PAPER_NAME
		name := windows.UTF16ToString(buffer[start : start+MAX_PAPER_NAME])
		if name != "" {
			sizes = append(sizes, name)
		}
	}
	return sizes
}

func GetPrinterList() []PrinterInfo {
	names, _ := printer.ReadNames()
	var printers []PrinterInfo
	for _, name := range names {
		id := HashPrinterName(name)
		printerMap[id] = name

		printers = append(printers, PrinterInfo{
			ID:         id,
			Name:       name,
			PaperSizes: GetSupportedPaperSizes(name),
		})
	}
	return printers
}

func GetPrinterMap() map[string]string {
	return printerMap
}

func GetPrinterInfo(id string) (PrinterInfo, bool) {
	name, ok := printerMap[id]
	if !ok {
		return PrinterInfo{}, false
	}
	return PrinterInfo{
		ID:         id,
		Name:       name,
		PaperSizes: GetSupportedPaperSizes(name),
	}, true
}
