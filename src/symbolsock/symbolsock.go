package symbolsock

import (
	"fmt"
	"encoding/json"
)

// This is a seriespoint representing an equity values are as follows: 
// [Open Price, Highest Price, Lowest Price, Closing Price]
type OHLC [4]float32

//This is like a Case Class in Scala, it represents the JSON object sent to clients
type seriesPoint struct {
	Symbol  string `json:"symb"`
	LastValue float32 `json:"lastVal"`
   	MinValue float32 `json:"min"`
	PointData OHLC `json:"data"`
}

// This will hold the sample data stored in Redis which all streams will
// FYI: All "streams" will use this one open since it's pretend Real Time Ticks
var dataPoints []OHLC

// This struct will represent a Ticker Feed such as Euros/USD, or JPY/GBP.
// atIndex is for storing the last index in dataPoints sent to a client
type symbolPush struct{
	symbol string
	clients int
	feed string
	lastV float32
	atIndex int
}

// each time a client subscribes to a feed, clients is incremented
func (s *symbolPush) increment() int {
	s.clients++
	fmt.Println(s.feed, " has ", s.clients, " listening")
	return s.clients
}

// each time a client subscribes to a feed, clients is decremented, if zero it will be deleted
func (s *symbolPush) decrement() int {
	s.clients--
	fmt.Println(s.feed, " has ", s.clients, " listening")
	return s.clients
}

//A function added to the symbolPush struct to return a Object literal which will be Stringified and broadcasted
func (s *symbolPush) GetPoint() seriesPoint {
	s.atIndex++
	lastVV := s.lastV
	if(s.atIndex == (len(dataPoints) -1)){
		s.atIndex = 0
	}
	//index 3 is the close value
	s.lastV = dataPoints[s.atIndex][3]
	minVal := dataPoints[s.atIndex][2]
    return seriesPoint{
    	Symbol: s.symbol,
    	LastValue: lastVV,
    	MinValue: minVal,
    	PointData: dataPoints[s.atIndex],
    }
}
type symbolHandler interface {
	    increment() int 
	    decrement() int
	    GetPoint() seriesPoint
}
//Create and returns a new symbolPush instance
func newSymbolPusher(sym string) symbolPush {
	feedName := sym + "_Feed"
	return symbolPush{
		symbol: sym,
		clients: 1,
		feed: feedName,
		lastV:  dataPoints[0][3],
		atIndex: 0,
	}
}

func parseJsonOb( blob string) {
	err := json.Unmarshal([]byte(blob), &dataPoints)
	if err != nil{
		panic(err)
	}
	fmt.Println(dataPoints[2])
	fmt.Println(len(dataPoints))
}
// closure function for managaing active streams
func SymbolStream( blob string) symbolHandler {
	parseJsonOb(blob)
	streamCount := 0
	runningStreams := make(map[string]symbolPush)
	
	newStream := func(syml string){
		elem, ok := runningStreams[syml]
		if ok{
			elem.increment()
		}else{
			streamCount++
			fmt.Println("new Symbol Stream created symbol is ", syml, " # of feeds " , streamCount)
			runningStreams[syml] = newSymbolPusher(syml)
		}
	}
	clientLeft := func(syml string) {
		elem, ok := runningStreams[syml]
		if ok{
			remain := elem.decrement()
			if(remain == 0){
			  streamCount--
			  fmt.Println("Symbol stream unsubscribed the symbol is ", syml, " # of feeds " , streamCount)
			  delete(runningStreams, syml)
			}
		}else{
			fmt.Println("Client said to leave a stream of NULL")
		}
	}

	getSeriesPoint := func(syml string) seriesPoint {
		elem, ok := runningStreams[syml]
		if ok {
		 return elem.GetPoint()
		}else{
			fmt.Println("Something Happened at getSeriesPoint in symbolHandler interface")
		}
       return elem.GetPoint()
	}
	

	
	return symbolHandler
} 