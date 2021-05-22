package app

import (
	"fmt"
	"math/rand"
	"time"
)

const DELAY = 2

func Sender(vertex *Vertex, print chan<- string, closingVariable *bool) {
	for {
		if *closingVariable {
			break
		}
		time.Sleep(time.Second * time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(DELAY)))

		reader := NewReaderPermission()
		vertex.ReaderLock.in <- reader
		<-reader.WatcherAccepted

		print <- fmt.Sprintf("\tSender %d: reading neighbours\n", vertex.Id)

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
			print <- fmt.Sprintf("\tSender %d: propagating packet  ", vertex.Id)
			for _, c := range packet.RoutingChanges {
				print <- fmt.Sprintf("{%d %d} ", c.SourceId, c.NewCost)
			}
			print <- " to : "
			for _, c := range vertex.NextVerticesInfo {
				print <- fmt.Sprintf("%d ", c.Id)
			}
			print <- "\n"
			for _, nextVertex := range vertex.NextVerticesInfo {
				nextVertex.PacketQueue.in <- CopyPacket(packet)
			}
		}

		print <- fmt.Sprintf("\tSender %d: ended work\n", vertex.Id)
		readerLeaving := <-reader.ReaderEnded
		readerLeaving <- true
	}
	close(vertex.PacketQueue.in)
}

func Receiver(vertex *Vertex, print chan<- string) {
	for changePacket := range vertex.PacketQueue.out {
		writer := NewWriterPermission()
		vertex.WriterLock.in <- writer
		<-writer.WatcherAccepted

		print <- fmt.Sprintf("\t\tReceiver %d: updating routing table as informed by %d\n", vertex.Id, changePacket.FromWho)

		for _, change := range changePacket.RoutingChanges {
			if change.SourceId == vertex.Id {
				continue
			}
			newCost := 1 + change.NewCost
			routingTable := vertex.ThisRoutingTable.RoutingInfos[change.SourceId]
			if newCost < routingTable.Cost {
				print <- fmt.Sprintf("\t\tReceiver %d: setting new route to %d (old: %d, new:%d) <-------------------------\n", vertex.Id, change.SourceId, routingTable.Cost, newCost)
				routingTable.Cost = newCost
				routingTable.NextHop = changePacket.FromWho
				routingTable.Changed = true
			} else {
				print <- fmt.Sprintf("\t\tReceiver %d: route to %d not changed (old: %d, new:%d)\n", vertex.Id, change.SourceId, routingTable.Cost, newCost)
			}
		}

		writer.WriterEnded <- true
	}
}

func Librarian(graph *Graph, print chan<- string, closingVariable *bool) {
	for {
		if *closingVariable {
			break
		}
		select {
		case <-graph.ReaderLeaving:
			print <- "Librarian: said bye to reader\n"
			graph.ReadersCount--
			break

		case writer := <-graph.WritersQueue.out:
			print <- "Librarian: welcomed writer\n"
			for {
				if graph.ReadersCount == 0 {
					print <- "Librarian: No readers left\n"
					break
				}
				print <- "Librarian: Waiting for readers to leave\n"
				<-graph.ReaderLeaving
				graph.ReadersCount--
				print <- "Librarian: said bye to reader\n"
			}
			print <- "Librarian: writer was let in\n"
			writer.WatcherAccepted <- true
			<-writer.WriterEnded
			print <- "Librarian: writer ended\n"
			break

		case reader := <-graph.ReadersQueue.out:
			print <- "Librarian: welcomed reader\n"
			graph.ReadersCount++
			print <- "Librarian: reader was let in\n"
			reader.WatcherAccepted <- true
			reader.ReaderEnded <- graph.ReaderLeaving
			break
		}
	}
	close(graph.ReaderLeaving)
	close(graph.WritersQueue.in)
	close(graph.ReadersQueue.in)
}
