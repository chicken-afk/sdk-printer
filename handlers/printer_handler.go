package handlers

import (
	"io"
	"net/http"

	"github.com/alexbrainman/printer"
	"github.com/gofiber/fiber/v2"
	"github.com/pos/sdk/utils"
)

type PrinterService interface {
	GetPrinters(c *fiber.Ctx) error
	GetPrinterPapers(c *fiber.Ctx) error
	Print(c *fiber.Ctx) error
}

type PrinterInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	PaperSizes []string `json:"paper_sizes"`
}

type printerServiceImpl struct{}

func NewPrinterService() PrinterService {
	return &printerServiceImpl{}
}

// Handler: GET /printers
func (s *printerServiceImpl) GetPrinters(c *fiber.Ctx) error {
	names, _ := printer.ReadNames()
	var printerMap = map[string]string{} // reset

	var printers []PrinterInfo
	for _, name := range names {
		id := utils.HashPrinterName(name)
		printerMap[id] = name

		printers = append(printers, PrinterInfo{
			ID:         id,
			Name:       name,
			PaperSizes: utils.GetSupportedPaperSizes(name),
		})
	}
	return c.JSON(printers)
}

// Handler: GET /printers/:id/papers
func (s *printerServiceImpl) GetPrinterPapers(c *fiber.Ctx) error {
	printerInfo := utils.GetPrinterList()

	//Validate id required
	if c.Params("id") == "" {
		return fiber.ErrBadRequest
	}

	var name string
	for _, p := range printerInfo {
		if p.ID == c.Params("id") {
			name = p.Name
			break
		}
	}

	papers := utils.GetSupportedPaperSizes(name)
	return c.JSON(papers)
}

// Handler: POST /print
func (s *printerServiceImpl) Print(c *fiber.Ctx) error {
	printerId := c.FormValue("printerId")
	paperSize := c.FormValue("paperSize")
	fileUrl := c.FormValue("fileUrl")

	var content []byte
	var err error

	// Cek apakah ada URL file
	if fileUrl != "" {
		// Download file dari URL
		resp, err := http.Get(fileUrl)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "failed to download file from URL: "+err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fiber.NewError(fiber.StatusBadRequest, "failed to download file, status code: "+resp.Status)
		}

		content, err = io.ReadAll(resp.Body)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read file from URL")
		}
	} else {
		// Cek apakah ada file upload
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "file or fileUrl is required")
		}
		file, err := fileHeader.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to open file")
		}
		defer file.Close()

		content, err = io.ReadAll(file)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read file")
		}
	}

	printerInfo := utils.GetPrinterList()
	var name string
	for _, p := range printerInfo {
		if p.ID == printerId {
			name = p.Name
			break
		}
	}

	if paperSize == "" {
		return fiber.NewError(fiber.StatusBadRequest, "paperSize is required")
	}

	h, err := printer.Open(name)
	if err != nil {
		return err
	}
	defer h.Close()
	// TODO: set paper size ke printer driver
	// applyPaperSize(h, paperSize)
	if err := h.StartDocument("PrintJob", "RAW"); err != nil {
		return err
	}
	if err := h.StartPage(); err != nil {
		return err
	}
	h.Write(content)
	h.EndPage()
	h.EndDocument()
	return nil
}
