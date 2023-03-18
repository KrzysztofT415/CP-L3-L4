package app

type Vertex struct {
	Id                  int
	ThisRoutingTable    *RoutingTable
	RoutingPacketQueue  *RoutingPacketQueue
	StandardPacketQueue *StandardPacketQueue
	Hosts				int
	HostsChannels       []chan *StandardPacket
	ReaderIn            bool
	ReaderLeaving       chan bool
	ReaderLock          chan *ReaderPermission
	WriterLock          chan *WriterPermission

	NextVerticesInfo []*Vertex
}

func NewVertex(id int, size int, hosts int) *Vertex {
	newVertex := new(Vertex)
	newVertex.Id = id
	newVertex.Hosts = hosts
	newVertex.HostsChannels = make([]chan *StandardPacket, hosts)
	for i := 0; i < hosts; i++ {
		newVertex.HostsChannels[i] = make(chan *StandardPacket)
	}
	newVertex.ThisRoutingTable = NewRoutingTable(size)
	newVertex.RoutingPacketQueue = MakeUnboundedQueueOfRoutingPackets()
	newVertex.StandardPacketQueue = MakeUnboundedQueueOfStandardPackets()
	newVertex.ReaderIn = false
	newVertex.ReaderLeaving = make(chan bool)
	newVertex.ReaderLock = make(chan *ReaderPermission)
	newVertex.WriterLock = make(chan *WriterPermission)
	return newVertex
}
