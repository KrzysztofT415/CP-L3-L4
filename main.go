package main

import (
	"PW-L3-GO/app"
	"fmt"
	"time"
)

const WORKINGTIME = 4

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
		close(printServerChannel)
		donePrinting <- true
	}()

	graph := app.NewGraph(n, d)

	printGraph(graph, printServerChannel)
	printRoutingTables(graph, printServerChannel)

	doneWorking := false

	printServerChannel <- "START OF TRANSFER\n\n"
	go app.Librarian(graph, printServerChannel, &doneWorking)
	for _, vertex := range graph.VerticesInfo {
		go app.Receiver(vertex, printServerChannel)
	}
	for _, vertex := range graph.VerticesInfo {
		go app.Sender(vertex, printServerChannel, &doneWorking)
	}

	<-time.After(time.Second * time.Duration(WORKINGTIME))
	doneWorking = true

	printServerChannel <- "\nEND OF TRANSFER\n---------------\n   LOG\n"
	printRoutingTables(graph, printServerChannel)

	printServerChannel <- "EOF"
	<-donePrinting
	close(donePrinting)
}

func printGraph(graph *app.Graph, print chan<- string) {
	print <- "---------------\n  GRAPH\nId:  Links:\n"
	for _, v := range graph.VerticesInfo {
		print <- fmt.Sprintf(" %d -> ", v.Id)
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
