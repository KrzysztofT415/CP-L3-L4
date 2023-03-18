package app

import (
	"fmt"
	"math/rand"
	"time"
)

const DELAY = 5

func RouterSender(vertex *Vertex, print chan<- string, closingVariable *bool) {
	for {
		if *closingVariable {
			break
		}
		time.Sleep(time.Second * time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(DELAY)))

		reader := NewReaderPermission()
		vertex.ReaderLock <- reader
		<-reader.WatcherAccepted

		print <- fmt.Sprintf("\tRouterSender %d: reading routing table\n", vertex.Id)

		packet := new(RoutingPacket)
		packet.FromWho = vertex.Id
		for _, routingInfo := range vertex.ThisRoutingTable.RoutingInfos {
			if routingInfo.Changed && routingInfo.SourceId != vertex.Id {
				routingInfo.Changed = false
				newChange := NewRoutingChange(routingInfo.SourceId, routingInfo.Cost)
				packet.RoutingChanges = append(packet.RoutingChanges, newChange)
			}
		}

		if len(packet.RoutingChanges) > 0 {
			info := ""
			for _, c := range packet.RoutingChanges {
				info += fmt.Sprintf("{%d %d} ", c.SourceId, c.NewCost)
			}
			info += " to : "
			for _, c := range vertex.NextVerticesInfo {
				info += fmt.Sprintf("%d ", c.Id)
			}

			print <- fmt.Sprintf("\tRouterSender %d: propagating packet  %s\n", vertex.Id, info)
			for _, nextVertex := range vertex.NextVerticesInfo {
				nextVertex.RoutingPacketQueue.in <- CopyRoutingPacket(packet)
			}
		}

		print <- fmt.Sprintf("\tRouterSender %d: ended work\n", vertex.Id)
		readerLeaving := <-reader.ReaderEnded
		readerLeaving <- true
	}
}

func RouterReceiver(vertex *Vertex, print chan<- string) {
	for changePacket := range vertex.RoutingPacketQueue.out {
		writer := NewWriterPermission()
		vertex.WriterLock <- writer
		<-writer.WatcherAccepted

		print <- fmt.Sprintf("\t\tRouterReceiver %d: updating routing table as informed by %d\n", vertex.Id, changePacket.FromWho)

		for _, change := range changePacket.RoutingChanges {
			if change.SourceId == vertex.Id {
				continue
			}
			newCost := 1 + change.NewCost
			routingTable := vertex.ThisRoutingTable.RoutingInfos[change.SourceId]
			if newCost < routingTable.Cost {
				print <- fmt.Sprintf("\t\tRouterReceiver %d: setting new route to %d (old: %d, new:%d) <-------------------------\n", vertex.Id, change.SourceId, routingTable.Cost, newCost)
				routingTable.Cost = newCost
				routingTable.NextHop = changePacket.FromWho
				routingTable.Changed = true
			} else {
				print <- fmt.Sprintf("\t\tRouterReceiver %d: route to %d not changed (old: %d, new:%d)\n", vertex.Id, change.SourceId, routingTable.Cost, newCost)
			}
		}

		writer.WriterEnded <- true
	}
}

func RouterForwarder(vertex *Vertex, print chan<- string) {
	for hostingPacket := range vertex.StandardPacketQueue.out {
		print <- fmt.Sprintf("\t\t\tRouterForwarder %d: received standard packet from %d-%d to %d-%d\n", vertex.Id, hostingPacket.FromRouter, hostingPacket.FromHost, hostingPacket.ToRouter, hostingPacket.ToHost)

		hostingPacket.VisitedRouters = append(hostingPacket.VisitedRouters, vertex.Id)

		if hostingPacket.ToRouter == vertex.Id {
			print <- fmt.Sprintf("\t\t\tRouterForwarder %d: giving to own host\n", vertex.Id)

			vertex.HostsChannels[hostingPacket.ToHost] <- hostingPacket

		} else {
			reader := NewReaderPermission()
			vertex.ReaderLock <- reader
			<-reader.WatcherAccepted

			nextHop := vertex.ThisRoutingTable.RoutingInfos[hostingPacket.ToRouter].NextHop
			print <- fmt.Sprintf("\t\t\tRouterForwarder %d: forwarding to nextHop: %d\n", vertex.Id, nextHop)

			for _, next := range vertex.NextVerticesInfo {
				if next.Id == nextHop {
					next.StandardPacketQueue.in <- hostingPacket
					break
				}
			}

			readerLeaving := <-reader.ReaderEnded
			readerLeaving <- true
		}

		print <- fmt.Sprintf("\t\t\tRouterForwarder %d: ended work\n", vertex.Id)
	}
}

func RouterReadWriteLock(vertex *Vertex, print chan<- string, closingVariable *bool) {
	for {
		if *closingVariable {
			break
		}
		select {
		case <-vertex.ReaderLeaving:
			print <- fmt.Sprintf("RouterReadWriteLock %d: said bye to reader\n", vertex.Id)
			vertex.ReaderIn = false
			break

		case writer := <-vertex.WriterLock:
			print <- fmt.Sprintf("RouterReadWriteLock %d: welcomed writer\n", vertex.Id)
			if vertex.ReaderIn {
				print <- fmt.Sprintf("RouterReadWriteLock %d: Waiting for reader to leave\n", vertex.Id)
				<-vertex.ReaderLeaving
				vertex.ReaderIn = false
				print <- fmt.Sprintf("RouterReadWriteLock %d: said bye to reader\n", vertex.Id)
			}
			print <- fmt.Sprintf("RouterReadWriteLock %d: writer was let in\n", vertex.Id)
			writer.WatcherAccepted <- true
			<-writer.WriterEnded
			print <- fmt.Sprintf("RouterReadWriteLock %d: writer ended\n", vertex.Id)
			break

		case reader := <-vertex.ReaderLock:
			print <- fmt.Sprintf("RouterReadWriteLock %d: welcomed reader\n", vertex.Id)
			vertex.ReaderIn = true
			print <- fmt.Sprintf("RouterReadWriteLock %d: reader was let in\n", vertex.Id)
			reader.WatcherAccepted <- true
			reader.ReaderEnded <- vertex.ReaderLeaving
			break
		}
	}
}

func HostWorker(vertex *Vertex, print chan<- string, closingVariable *bool, thisHostId int, destinationHost int, destinationRouter int) {
	time.Sleep(time.Millisecond * time.Duration(100))
	print <- fmt.Sprintf("\tHostWorker %d-%d: sending packet to %d-%d\n", vertex.Id, thisHostId, destinationRouter, destinationHost)

	packet := new(StandardPacket)
	packet.FromHost = thisHostId
	packet.FromRouter = vertex.Id
	packet.ToHost = destinationHost
	packet.ToRouter = destinationRouter
	vertex.StandardPacketQueue.in <- packet

	for {
		if *closingVariable {
			break
		}
		print <- fmt.Sprintf("\tHostWorker %d-%d: waiting for packet\n", vertex.Id, thisHostId)
		packetR := <-vertex.HostsChannels[thisHostId]
		path := ""
		for _, r := range packetR.VisitedRouters {
			path += fmt.Sprintf("%d ", r)
		}
		print <- fmt.Sprintf("\tHostWorker %d-%d: got packet | from %d-%d to %d-%d = %s\n", vertex.Id, thisHostId, packetR.FromRouter, packetR.FromHost, packetR.ToRouter, packetR.ToHost, path)

		time.Sleep(time.Second * time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(DELAY)))
		print <- fmt.Sprintf("\tHostWorker %d-%d: sending response to h:%d r:%d\n", vertex.Id, thisHostId, packetR.FromRouter, packetR.FromHost)

		packetS := new(StandardPacket)
		packetS.FromHost = thisHostId
		packetS.FromRouter = vertex.Id
		packetS.ToHost = packetR.FromHost
		packetS.ToRouter = packetR.FromRouter
		vertex.StandardPacketQueue.in <- packetS
	}
}
