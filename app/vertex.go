package app

type Vertex struct {
	Id               int
	ThisRoutingTable *RoutingTable
	PacketQueue      *PacketQueue
	ReaderLock       *ReaderQueue
	WriterLock       *WriterQueue

	NextVerticesInfo []*Vertex
}

func NewVertex(id int, size int, readersQueue *ReaderQueue, writersQueue *WriterQueue) *Vertex {
	newVertex := new(Vertex)
	newVertex.Id = id
	newVertex.ThisRoutingTable = NewRoutingTable(size)
	newVertex.PacketQueue = MakeUnboundedQueueOfPackets()
	newVertex.ReaderLock = readersQueue
	newVertex.WriterLock = writersQueue
	return newVertex
}
