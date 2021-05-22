package app

type RoutingInfo struct {
	SourceId int
	NextHop  int
	Cost     int
	Changed  bool
}

func NewRoutingInfo(sourceId int, nextHop int, cost int) *RoutingInfo {
	newRoutingInfo := new(RoutingInfo)
	newRoutingInfo.SourceId = sourceId
	newRoutingInfo.NextHop = nextHop
	newRoutingInfo.Cost = cost
	newRoutingInfo.Changed = true
	return newRoutingInfo
}

type RoutingTable struct {
	RoutingInfos []*RoutingInfo
}

func NewRoutingTable(size int) *RoutingTable {
	newRoutingTable := new(RoutingTable)
	newRoutingTable.RoutingInfos = make([]*RoutingInfo, size)
	return newRoutingTable
}

type RoutingChange struct {
	SourceId int
	NewCost  int
}

func NewRoutingChange(sourceId int, newCost int) *RoutingChange {
	newRoutingChange := new(RoutingChange)
	newRoutingChange.SourceId = sourceId
	newRoutingChange.NewCost = newCost
	return newRoutingChange
}

type RoutingPacket struct {
	FromWho        int
	RoutingChanges []*RoutingChange
}

func CopyPacket(packet *RoutingPacket) *RoutingPacket {
	copyPacket := new(RoutingPacket)
	copyPacket.FromWho = packet.FromWho
	for _, change := range packet.RoutingChanges {
		newRoutingChange := new(RoutingChange)
		newRoutingChange.SourceId = change.SourceId
		newRoutingChange.NewCost = change.NewCost
		copyPacket.RoutingChanges = append(copyPacket.RoutingChanges, newRoutingChange)
	}
	return copyPacket
}

type PacketQueue struct {
	in  chan<- *RoutingPacket
	out <-chan *RoutingPacket
}

func MakeUnboundedQueueOfPackets() *PacketQueue {
	in := make(chan *RoutingPacket)
	out := make(chan *RoutingPacket)

	go func() {
		var inQueue []*RoutingPacket

		outCh := func() chan *RoutingPacket {
			if len(inQueue) == 0 {
				return nil
			}

			return out
		}

		cur := func() *RoutingPacket {
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

	newPacketQueue := new(PacketQueue)
	newPacketQueue.in = in
	newPacketQueue.out = out
	return newPacketQueue
}
