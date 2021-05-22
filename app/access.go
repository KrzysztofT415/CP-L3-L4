package app

type WriterPermission struct {
	WatcherAccepted chan bool
	WriterEnded     chan bool
}

func NewWriterPermission() *WriterPermission {
	newWriterPermission := new(WriterPermission)
	newWriterPermission.WatcherAccepted = make(chan bool, 0)
	newWriterPermission.WriterEnded = make(chan bool, 0)
	return newWriterPermission
}

type WriterQueue struct {
	in  chan<- *WriterPermission
	out <-chan *WriterPermission
}

type ReaderPermission struct {
	WatcherAccepted chan bool
	ReaderEnded     chan chan bool
}

func NewReaderPermission() *ReaderPermission {
	newReaderPermission := new(ReaderPermission)
	newReaderPermission.WatcherAccepted = make(chan bool, 0)
	newReaderPermission.ReaderEnded = make(chan chan bool, 0)
	return newReaderPermission
}

type ReaderQueue struct {
	in  chan<- *ReaderPermission
	out <-chan *ReaderPermission
}

func MakeUnboundedQueueOfWriters() *WriterQueue {
	in := make(chan *WriterPermission)
	out := make(chan *WriterPermission)

	go func() {
		var inQueue []*WriterPermission

		outCh := func() chan *WriterPermission {
			if len(inQueue) == 0 {
				return nil
			}

			return out
		}

		cur := func() *WriterPermission {
			if len(inQueue) == 0 {
				return nil
			}

			return inQueue[0]
		}

		for len(inQueue) > 0 || in != nil {
			select {
			case oc, ok := <-in:
				if !ok {
					in = nil
				} else {
					inQueue = append(inQueue, oc)
				}
			case outCh() <- cur():
				if out != nil {
					inQueue = inQueue[1:]
				}
			}
		}

		close(out)
	}()

	newWriterQueue := new(WriterQueue)
	newWriterQueue.in = in
	newWriterQueue.out = out
	return newWriterQueue
}

func MakeUnboundedQueueOfReaders() *ReaderQueue {
	in := make(chan *ReaderPermission)
	out := make(chan *ReaderPermission)

	go func() {
		var inQueue []*ReaderPermission

		outCh := func() chan *ReaderPermission {
			if len(inQueue) == 0 {
				return nil
			}

			return out
		}

		cur := func() *ReaderPermission {
			if len(inQueue) == 0 {
				return nil
			}

			return inQueue[0]
		}

		for len(inQueue) > 0 || in != nil {
			select {
			case oc, ok := <-in:
				if !ok {
					in = nil
				} else {
					inQueue = append(inQueue, oc)
				}
			case outCh() <- cur():
				if out != nil {
					inQueue = inQueue[1:]
				}
			}
		}

		close(out)
	}()

	newReaderQueue := new(ReaderQueue)
	newReaderQueue.in = in
	newReaderQueue.out = out
	return newReaderQueue
}
