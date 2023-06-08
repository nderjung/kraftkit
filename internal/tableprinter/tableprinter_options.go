package tableprinter

type TableOutputFormat string

const (
	OutputFormatTable = TableOutputFormat("table")
	OutputFormatJSON  = TableOutputFormat("json")
	OutputFormatYAML  = TableOutputFormat("yaml")
)

type tablePrinterOption func(*TablePrinterOptions)

func WithOutputFormat(format TableOutputFormat) tablePrinterOption {
	return func(opts *TablePrinterOptions) {
		opts.format = format
	}
}

func WithIsTTY(istty bool) tablePrinterOption {
	return func(opts *TablePrinterOptions) {
		opts.IsTTY = istty
	}
}
