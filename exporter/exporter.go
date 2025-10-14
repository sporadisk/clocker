package exporter

type Exporter interface {
	Export(events []Event) error
}
