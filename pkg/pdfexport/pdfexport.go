package pdfexport

type PdfExport interface {
	Save(srcFile string) error
}
