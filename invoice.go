package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

const defaultFromName = "Your Company Inc"
const defaultFromAddress = "Unit 1, Lingkaran Syed Putra, Mid Valley City, 59200 Kuala Lumpur, Wilayah Persekutuan Kuala Lumpur"
const defaultFromContact = "03-9876 5432"
const defaultToName = "Target Company Inc"
const defaultToAddress = "Unit 999, Lingkaran Syed Putra, Mid Valley City, 59200 Kuala Lumpur, Wilayah Persekutuan Kuala Lumpur"
const defaultToContact = "03-1234 5678"

// Input flags
var invoiceNo string
var invoiceDate string
var companyNo string
var fromName string
var fromAddress string
var fromContact string
var toName string
var toAddress string
var toContact string
var taxPercent int

var invoiceCmd = &cobra.Command{
	Use:   "generate [CSV file]",
	Short: "Generate invoice from CSV file containing the items for the invoice",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		csvFilePath, err := filepath.Abs(args[0])
		must(err)

		// Check if csv file exists
		if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
			must(fmt.Errorf("csv file %v does not exists", csvFilePath))
		}

		// Check if the file is indeed csv file
		if filepath.Ext(csvFilePath) != ".csv" {
			must(fmt.Errorf("invalid file type: %v, only accept .csv file", filepath.Ext(csvFilePath)))
		}

		// Read data from csv
		data, err := readDataFromCSV(csvFilePath)
		must(err)

		must(generateInvoice(data))
	},
}

func init() {
	invoiceCmd.Flags().StringVarP(&invoiceNo, "invoiceNo", "n", "", "Invoice No., default: <empty>")
	invoiceCmd.Flags().StringVarP(&invoiceDate, "invoiceDate", "d", time.Now().Local().Format("2006-01-02"), "Invoice date in the format of YYYY-MM-DD, default: today's date")
	invoiceCmd.Flags().StringVarP(&companyNo, "companyNo", "p", "", "Company No., default: <empty>")
	invoiceCmd.Flags().StringVarP(&fromName, "fromName", "f", defaultFromName, fmt.Sprintf("From name, default: %v", defaultFromName))
	invoiceCmd.Flags().StringVarP(&fromAddress, "fromAddress", "a", defaultFromAddress, fmt.Sprintf("From address, default: %v", defaultFromAddress))
	invoiceCmd.Flags().StringVarP(&fromContact, "fromContact", "c", defaultFromContact, fmt.Sprintf("From contact, default: %v", defaultFromContact))
	invoiceCmd.Flags().StringVarP(&toName, "toName", "o", defaultToName, fmt.Sprintf("To name, default: %v", defaultToName))
	invoiceCmd.Flags().StringVarP(&toAddress, "toAddress", "r", defaultToAddress, fmt.Sprintf("To address, default: %v", defaultToAddress))
	invoiceCmd.Flags().StringVarP(&toContact, "toContact", "t", defaultToContact, fmt.Sprintf("To contact, default: %v", defaultToContact))
	invoiceCmd.Flags().IntVarP(&taxPercent, "taxPercent", "e", 5, "Tax percentage, default: 5%")

	rootCmd.AddCommand(invoiceCmd)
}

func generateInvoice(data [][]string) error {
	marginX := 10.0
	marginY := 20.0
	//gapX := float64(2)
	gapY := 2.0
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(marginX, marginY, marginX)
	pageW, _ := pdf.GetPageSize()
	safeAreaW := pageW - 2*marginX

	pdf.ImageOptions("assets/logo.png", 0, 0, 65, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.SetFont("Arial", "B", 16)
	_, lineHeight := pdf.GetFontSize()
	currentY := pdf.GetY() + lineHeight + gapY
	pdf.SetXY(marginX, currentY)
	pdf.Cell(40, 10, fromName)

	if companyNo != "" {
		pdf.SetFont("Arial", "BI", 12)
		_, lineHeight = pdf.GetFontSize()
		pdf.SetXY(marginX, pdf.GetY()+lineHeight+gapY)
		pdf.Cell(40, 10, fmt.Sprintf("Company No : %v", companyNo))
	}

	leftY := pdf.GetY() + lineHeight + gapY
	// Build invoice word on right
	pdf.SetFont("Arial", "B", 32)
	_, lineHeight = pdf.GetFontSize()
	pdf.SetXY(130, currentY-lineHeight)
	pdf.Cell(100, 40, "INVOICE")

	newY := leftY
	if (pdf.GetY() + gapY) > newY {
		newY = pdf.GetY() + gapY
	}

	newY += 10.0 // Add margin

	pdf.SetXY(marginX, newY)
	pdf.SetFont("Arial", "", 12)
	_, lineHeight = pdf.GetFontSize()
	lineBreak := lineHeight + float64(1)

	// Left hand info
	splittedFromAddress := breakAddress(fromAddress)
	for _, add := range splittedFromAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf("Tel: %s", fromContact))
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)

	pdf.SetFontStyle("B")
	pdf.Cell(safeAreaW/2, lineHeight, "Bill To:")
	pdf.Line(marginX, pdf.GetY()+lineHeight, marginX+safeAreaW/2, pdf.GetY()+lineHeight)
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW/2, lineHeight, toName)
	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	splittedToAddress := breakAddress(toAddress)
	for _, add := range splittedToAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf("Tel: %s", toContact))

	endOfInvoiceDetailY := pdf.GetY() + lineHeight
	pdf.SetFontStyle("")

	// Right hand side info, invoice no & invoice date
	invoiceDetailW := float64(30)
	pdf.SetXY(safeAreaW/2+30, newY)
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice No.:")
	pdf.Cell(invoiceDetailW, lineHeight, invoiceNo)
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice Date:")
	pdf.Cell(invoiceDetailW, lineHeight, invoiceDate)
	pdf.Ln(lineBreak)

	// Draw the table
	pdf.SetXY(marginX, endOfInvoiceDetailY+10.0)
	lineHt := 10.0
	const colNumber = 5
	header := [colNumber]string{"No", "Description", "Quantity", "Unit Price ($)", "Price ($)"}
	colWidth := [colNumber]float64{10.0, 75.0, 25.0, 40.0, 40.0}

	// Headers
	pdf.SetFontStyle("B")
	pdf.SetFillColor(200, 200, 200)
	for colJ := 0; colJ < colNumber; colJ++ {
		pdf.CellFormat(colWidth[colJ], lineHt, header[colJ], "1", 0, "CM", true, 0, "")
	}

	pdf.Ln(-1)
	pdf.SetFillColor(255, 255, 255)

	// Table data
	pdf.SetFontStyle("")
	subtotal := 0.0

	for rowJ := 0; rowJ < len(data); rowJ++ {
		val := data[rowJ]
		if len(val) == 3 {
			// Column 1: Unit
			// Column 2: Description
			// Column 3: Price per unit
			unit, _ := strconv.Atoi(val[0])
			desc := val[1]
			pricePerUnit, _ := strconv.ParseFloat(val[2], 64)
			pricePerUnit = math.Round(pricePerUnit*100) / 100
			totalPrice := float64(unit) * pricePerUnit
			subtotal += totalPrice

			pdf.CellFormat(colWidth[0], lineHt, fmt.Sprintf("%d", rowJ+1), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[1], lineHt, desc, "1", 0, "LM", true, 0, "")
			pdf.CellFormat(colWidth[2], lineHt, fmt.Sprintf("%d", unit), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[3], lineHt, fmt.Sprintf("%.2f", pricePerUnit), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", totalPrice), "1", 0, "CM", true, 0, "")
			pdf.Ln(-1)
		}
	}

	// Calculate the subtotal
	pdf.SetFontStyle("B")
	leftIndent := 0.0
	for i := 0; i < 3; i++ {
		leftIndent += colWidth[i]
	}
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Subtotal", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", subtotal), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	taxAmount := math.Round(subtotal*float64(taxPercent)) / 100
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Tax Amount", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", taxAmount), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	grandTotal := subtotal + taxAmount
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Grand total", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", grandTotal), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, "Note: The tax invoice is computer generated and no signature is required.")

	return pdf.OutputFileAndClose("invoice.pdf")
}

func breakAddress(input string) []string {
	var address []string
	const limit = 10
	splitted := strings.Split(input, ",")
	prevAddress := ""
	for _, add := range splitted {
		if len(add) < 10 {
			prevAddress = add
			continue
		}
		currentAdd := strings.TrimSpace(add)
		if prevAddress != "" {
			currentAdd = prevAddress + ", " + currentAdd
		}
		address = append(address, currentAdd)
		prevAddress = ""
	}

	return address
}
