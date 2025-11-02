package event

type Exporter interface {
	Export(events []*Event) error
}
