package utils

import (
	"encoding/base64"
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
	//Change name to base64
	printerId := base64.StdEncoding.EncodeToString([]byte(name))
	return printerId
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

// Windows printer status constants
const (
	PRINTER_STATUS_READY             = 0x00000000
	PRINTER_STATUS_PAUSED            = 0x00000001
	PRINTER_STATUS_ERROR             = 0x00000002
	PRINTER_STATUS_PENDING_DELETION  = 0x00000004
	PRINTER_STATUS_PAPER_JAM         = 0x00000008
	PRINTER_STATUS_PAPER_OUT         = 0x00000010
	PRINTER_STATUS_MANUAL_FEED       = 0x00000020
	PRINTER_STATUS_PAPER_PROBLEM     = 0x00000040
	PRINTER_STATUS_OFFLINE           = 0x00000080
	PRINTER_STATUS_IO_ACTIVE         = 0x00000100
	PRINTER_STATUS_BUSY              = 0x00000200
	PRINTER_STATUS_PRINTING          = 0x00000400
	PRINTER_STATUS_OUTPUT_BIN_FULL   = 0x00000800
	PRINTER_STATUS_NOT_AVAILABLE     = 0x00001000
	PRINTER_STATUS_WAITING           = 0x00002000
	PRINTER_STATUS_PROCESSING        = 0x00004000
	PRINTER_STATUS_INITIALIZING      = 0x00008000
	PRINTER_STATUS_WARMING_UP        = 0x00010000
	PRINTER_STATUS_TONER_LOW         = 0x00020000
	PRINTER_STATUS_NO_TONER          = 0x00040000
	PRINTER_STATUS_PAGE_PUNT         = 0x00080000
	PRINTER_STATUS_USER_INTERVENTION = 0x00100000
	PRINTER_STATUS_OUT_OF_MEMORY     = 0x00200000
	PRINTER_STATUS_DOOR_OPEN         = 0x00400000
	PRINTER_STATUS_SERVER_UNKNOWN    = 0x00800000
	PRINTER_STATUS_POWER_SAVE        = 0x01000000
)

// PRINTER_INFO_2 structure for getting detailed printer information
type PRINTER_INFO_2 struct {
	pServerName         *uint16
	pPrinterName        *uint16
	pShareName          *uint16
	pPortName           *uint16
	pDriverName         *uint16
	pComment            *uint16
	pLocation           *uint16
	pDevMode            uintptr
	pSepFile            *uint16
	pPrintProcessor     *uint16
	pDatatype           *uint16
	pParameters         *uint16
	pSecurityDescriptor uintptr
	Attributes          uint32
	Priority            uint32
	DefaultPriority     uint32
	StartTime           uint32
	UntilTime           uint32
	Status              uint32
	cJobs               uint32
	AveragePPM          uint32
}

// GetPrinterStatus returns the current status of a printer
func GetPrinterStatus(printerName string) string {
	winspool := windows.NewLazySystemDLL("winspool.drv")
	openPrinter := winspool.NewProc("OpenPrinterW")
	closePrinter := winspool.NewProc("ClosePrinter")
	getPrinter := winspool.NewProc("GetPrinterW")

	var hPrinter uintptr
	printerNamePtr := windows.StringToUTF16Ptr(printerName)

	// Open printer
	ret, _, _ := openPrinter.Call(
		uintptr(unsafe.Pointer(printerNamePtr)),
		uintptr(unsafe.Pointer(&hPrinter)),
		0,
	)

	if ret == 0 {
		return "offline" // Cannot open printer
	}
	defer closePrinter.Call(hPrinter)

	// Get required buffer size
	var bytesNeeded uint32
	getPrinter.Call(
		hPrinter,
		2, // Level 2 for PRINTER_INFO_2
		0,
		0,
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)

	if bytesNeeded == 0 {
		return "unknown"
	}

	// Allocate buffer and get printer info
	buffer := make([]byte, bytesNeeded)
	ret, _, _ = getPrinter.Call(
		hPrinter,
		2,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bytesNeeded),
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)

	if ret == 0 {
		return "unknown"
	}

	// Parse PRINTER_INFO_2
	info := (*PRINTER_INFO_2)(unsafe.Pointer(&buffer[0]))
	status := info.Status

	// Check status flags and return appropriate status
	if status&PRINTER_STATUS_OFFLINE != 0 {
		return "offline"
	}
	if status&PRINTER_STATUS_ERROR != 0 {
		return "error"
	}
	if status&PRINTER_STATUS_PAPER_JAM != 0 {
		return "paper_jam"
	}
	if status&PRINTER_STATUS_PAPER_OUT != 0 {
		return "paper_out"
	}
	if status&PRINTER_STATUS_PAPER_PROBLEM != 0 {
		return "paper_problem"
	}
	if status&PRINTER_STATUS_DOOR_OPEN != 0 {
		return "door_open"
	}
	if status&PRINTER_STATUS_NO_TONER != 0 {
		return "no_toner"
	}
	if status&PRINTER_STATUS_TONER_LOW != 0 {
		return "toner_low"
	}
	if status&PRINTER_STATUS_OUT_OF_MEMORY != 0 {
		return "out_of_memory"
	}
	if status&PRINTER_STATUS_USER_INTERVENTION != 0 {
		return "user_intervention"
	}
	if status&PRINTER_STATUS_OUTPUT_BIN_FULL != 0 {
		return "output_bin_full"
	}
	if status&PRINTER_STATUS_NOT_AVAILABLE != 0 {
		return "not_available"
	}
	if status&PRINTER_STATUS_PAUSED != 0 {
		return "paused"
	}
	if status&PRINTER_STATUS_PRINTING != 0 {
		return "printing"
	}
	if status&PRINTER_STATUS_BUSY != 0 {
		return "busy"
	}
	if status&PRINTER_STATUS_PROCESSING != 0 {
		return "processing"
	}
	if status&PRINTER_STATUS_INITIALIZING != 0 {
		return "initializing"
	}
	if status&PRINTER_STATUS_WARMING_UP != 0 {
		return "warming_up"
	}
	if status&PRINTER_STATUS_POWER_SAVE != 0 {
		return "power_save"
	}
	if status&PRINTER_STATUS_WAITING != 0 {
		return "waiting"
	}

	// If no status flags set, printer is ready/idle
	return "idle"
}
