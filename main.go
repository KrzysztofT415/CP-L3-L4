package main

import (
	"fmt"
	"math/rand"
	"time"

	"./app"
)

const WORKINGTIME = 6

func main() {
	//n - vertices, d - shortcuts
	n, d := 4, 1

	printServerChannel := make(chan string, 50)
	donePrinting := make(chan bool)
	go func() {
		for info := range printServerChannel {
			if info == "EOF" {
				break
			}
			fmt.Print(info)
		}
		donePrinting <- true
	}()

	silentServerChannel := make(chan string, 50)
	go func() {
		for range silentServerChannel {}
	}()

	graph := app.NewGraph(n, d)

	printGraph(graph, printServerChannel)
	printRoutingTables(graph, printServerChannel)

	doneWorking := false

	printServerChannel <- "START OF TRANSFER\n\n"
	for _, vertex := range graph.VerticesInfo {
		go app.RouterReadWriteLock(vertex, silentServerChannel, &doneWorking)
		go app.RouterReceiver(vertex, silentServerChannel)
	}
	for _, vertex := range graph.VerticesInfo {
		go app.RouterSender(vertex, silentServerChannel, &doneWorking)
		go app.RouterForwarder(vertex, printServerChannel)
	}
	for _, vertex := range graph.VerticesInfo {
		for j := 0; j < vertex.Hosts; j++ {
			rand.Seed(time.Now().UnixNano())
			time.Sleep(time.Microsecond * time.Duration(10))
			for {
				destR := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(graph.VerticesInfo))
				if destR == vertex.Id { continue }

				destH := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(graph.VerticesInfo[destR].Hosts)
				go app.HostWorker(vertex, printServerChannel, &doneWorking, j, destH, destR)
				break
			}
		}
	}

	<-time.After(time.Second * time.Duration(WORKINGTIME))
	doneWorking = true

	printServerChannel <- "\nEND OF TRANSFER\n---------------\n   LOG\n"
	printRoutingTables(graph, printServerChannel)

	printServerChannel <- "EOF"
	<-donePrinting
}

func printGraph(graph *app.Graph, print chan<- string) {
	print <- "---------------\n  GRAPH\nHosts: Id:  Links:\n"
	for _, v := range graph.VerticesInfo {
		print <- fmt.Sprintf("%d   %d -> ", v.Hosts, v.Id)
		for _, c := range v.NextVerticesInfo {
			print <- fmt.Sprintf("%d ", c.Id)
		}
		print <- "\n"
	}
	print <- fmt.Sprintf("---------------\n")
}

func printRoutingTables(graph *app.Graph, print chan<- string) {
	print <- "  ROUTING TABLES\nId:  Routes {(id),nextHop,cost,changed}:\n"
	for _, v := range graph.VerticesInfo {
		print <- fmt.Sprintf(" %d := ", v.Id)
		for _, c := range v.ThisRoutingTable.RoutingInfos {
			changed := "F"
			if c.Changed {
				changed = "T"
			}
			if c.NextHop == -1 {
				print <- fmt.Sprintf("{    -    } ")
			} else {
				print <- fmt.Sprintf("{(%d) %d %d %s} ", c.SourceId, c.NextHop, c.Cost, changed)
			}
		}
		print <- "\n"
	}
	print <- fmt.Sprintf("---------------\n")
}
