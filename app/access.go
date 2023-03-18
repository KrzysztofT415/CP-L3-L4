package app

type WriterPermission struct {
	WatcherAccepted chan bool
	WriterEnded     chan bool
}

func NewWriterPermission() *WriterPermission {
	newWriterPermission := new(WriterPermission)
	newWriterPermission.WatcherAccepted = make(chan bool)
	newWriterPermission.WriterEnded = make(chan bool)
	return newWriterPermission
}

type ReaderPermission struct {
	WatcherAccepted chan bool
	ReaderEnded     chan chan bool
}

func NewReaderPermission() *ReaderPermission {
	newReaderPermission := new(ReaderPermission)
	newReaderPermission.WatcherAccepted = make(chan bool)
	newReaderPermission.ReaderEnded = make(chan chan bool)
	return newReaderPermission
}
