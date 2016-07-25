package ticksimulator

import (
  "fmt"
  "math/rand"
  "time"
)
// This is a seriespoint representing an equity values are as follows: 
// [Open Price, Highest Price, Lowest Price, Closing Price]
type OHLC [4]float32

//2016 Yearly High and Low of different currencies plus the index[0] is the "last" price
type LastHighLow [3]float32

//Case Class, for JSON object sent to clients
type SeriesPoint struct {
	MaxYear float32 `json:"maxyear"`
	MinYear float32 `json:"minyear"`
	Symbol  string `json:"symb"`
	LastValue float32 `json:"lastVal"`
   	MinValue float32 `json:"min"`
	PointData OHLC `json:"data"`
}
//Symbol push holds state such as number of listeners and the last value.
type SymbolPush struct{
	symbol string
	feed string
	clients int
	highest float32
	lowest float32
	lastV float32
}

// each time a client subscribes to a feed, clients is incremented
func (s *SymbolPush) increment() int {
	s.clients++
	fmt.Println(s.feed, " has ", s.clients, " listening")
	return s.clients
}

// each time a client subscribes to a feed, clients is decremented, if zero it will be deleted
func (s *SymbolPush) decrement() int {
	s.clients--
	fmt.Println(s.feed, " has ", s.clients, " listening")
	return s.clients
}

//caluclates a fake Tick Point using the last price and a range between the yearly Highest and Lowest
func getOHLCpoint(sym string, open float32, highest float32, lowest float32) OHLC {
    r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
    g := r1.Float32()
    low := open - ((highest - lowest) * g )
    high := open +  ((highest - lowest) * g )
    clse := low + (r1.Float32() * (high - low ))
   // fmt.Println( highest ," then ", lowest, " rand = ",  low, high, clse)

    return OHLC{open, high, low, clse}
}

//this calculates a new Point and returns the SeriesPoint case class
func (s *SymbolPush) GetPoint() SeriesPoint {
	lastVV := s.lastV
    ohlc := getOHLCpoint( s.symbol, s.lastV, s.highest, s.lowest) 
	//index 3 is the close value
	s.lastV = ohlc[3]
	minVal := ohlc[2]
    return SeriesPoint{
    	MaxYear: s.highest,
    	MinYear: s.lowest,
    	Symbol: s.symbol,
    	LastValue: lastVV,
    	MinValue: minVal,
    	PointData: ohlc,
    }
}


var SymStreams = map[string]SymbolPush{ 
	  "AUD/JPY": { "AUD/JPY", "AUD/JPY_feed", 0, 79.013, 87.815, 72.354 },
	  "AUD/USD": { "AUD/USD", "AUD/USD_feed", 0, 0.74521, 0.78337, 0.68265 },
	  "CAD/CHF": { "CAD/CHF", "CAD/CHF_feed", 0, 0.7514, 0.7735, 0.6803 },
	  "CAD/JPY": { "CAD/JPY", "CAD/JPY_feed", 0, 80.757, 88.796, 76.071 },
	  "CHF/JPY": { "CHF/JPY", "CHF/JPY_feed", 0, 107.342, 120.365, 101.195 },
	  "EUR/AUD": { "EUR/AUD", "EUR/AUD_feed", 0, 1.46873, 1.62429, 1.44335 },
	  "EUR/CAD": { "EUR/CAD", "EUR/CAD_feed", 0, 1.44065, 1.61045, 1.41772 },
	  "EUR/GBP": { "EUR/GBP", "EUR/GBP_feed", 0, 0.83713, 0.86255, 0.73119 },
	  "EUR/JPY": { "EUR/JPY", "EUR/JPY_feed", 0, 116.393, 132.285, 109.552 },
	  "EUR/NOK": { "EUR/NOK", "EUR/NOK_feed", 0, 9.37867, 9.74633, 9.15502 },
	  "EUR/SEK": { "EUR/SEK", "EUR/SEK_feed", 0, 9.49165, 9.61074, 9.11289 },
	  "EUR/USD": { "EUR/USD", "EUR/USD_feed", 0, 1.09758, 1.1616, 1.07102 },
	  "GBP/AUD": { "GBP/AUD", "GBP/AUD_feed", 0, 1.75303, 2.09769, 1.7043 },
	  "GBP/CHF": { "GBP/CHF", "GBP/CHF_feed", 0, 1.2927, 1.48345, 1.25059 },
	  "GBP/JPY": { "GBP/JPY", "GBP/JPY_feed", 0, 138.921, 177.356, 128.788 },
	  "GBP/USD": { "GBP/USD", "GBP/USD_feed", 0, 1.31064, 1.50155, 1.27975 },
	  "NZD/JPY": { "NZD/JPY", "NZD/JPY_feed", 0, 74.079, 82.23, 68.889 },
	  "NZD/USD": { "NZD/USD", "NZD/USD_feed", 0, 0.6988, 0.73236, 0.63463 },
	  "USD/BRL": { "USD/BRL", "USD/BRL_feed", 0, 3.25397, 4.16152, 3.18546 },
	  "USD/CAD": { "USD/CAD", "USD/CAD_feed", 0, 1.31252, 1.46896, 1.24602 },
	  "USD/CHF": { "USD/CHF", "USD/CHF_feed", 0, 0.98669, 1.02558, 0.94439 },
	  "USD/CNY": { "USD/CNY", "USD/CNY_feed", 0, 6.6758, 6.7134, 6.4108 },
	  "USD/CZK": { "USD/CZK", "USD/CZK_feed", 0, 24.6089, 25.2263, 23.2712 },
	  "USD/HKD": { "USD/HKD", "USD/HKD_feed", 0, 7.7566, 7.829, 7.7493 },
	  "USD/INR": { "USD/INR", "USD/INR_feed", 0, 67.0958, 68.8808, 65.88 },
	  "USD/JPY": { "USD/JPY", "USD/JPY_feed", 0, 106.033, 121.676, 99.075 },
	  "USD/KRW": { "USD/KRW", "USD/KRW_feed", 0, 1134.52, 1244.1, 1122.87 },
	  "USD/MXN": { "USD/MXN", "USD/MXN_feed", 0, 18.5388, 19.5088, 17.0489 },
	  "USD/NOK": { "USD/NOK", "USD/NOK_feed", 0, 8.54192, 8.99242, 7.96395 },
	  "USD/PLN": { "USD/PLN", "USD/PLN_feed", 0, 3.9742, 4.1552, 3.7043 },
	  "USD/RUB": { "USD/RUB", "USD/RUB_feed", 0, 64.745, 85.099, 62.791 },
	  "USD/SEK": { "USD/SEK", "USD/SEK_feed", 0, 8.64657, 8.74226, 7.89232 },
	  "USD/SGD": { "USD/SGD", "USD/SGD_feed", 0, 1.3579, 1.4443, 1.3311 },
	  "USD/ZAR": { "USD/ZAR", "USD/ZAR_feed", 0, 14.27189, 17.53799, 14.11169 }, 
}

func GetTickStruct(symb string) SymbolPush {
	return SymStreams[symb]
}
