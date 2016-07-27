package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"time"
	"math/rand"
    "websocket"
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}
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
type Client struct {
	ws   *websocket.Conn
	send chan SeriesPoint
	subs []string
} 
//Symbol push holds state such as number of listeners and the last value.
type SymbolPush struct{
	symbol string
	feed string
	clients int
	highest float32
	lowest float32
	lastV float32
	sockets map[*Client]bool
	addClient    chan *Client
	removeClient chan *Client
}
func (sb *SymbolPush) start() {
	tick := time.Tick(1 *time.Second)
	for {
		select {
		case conn := <- sb.addClient:
			sb.sockets[conn] = true
		case conn := <- sb.removeClient:
			if _, ok := sb.sockets[conn]; ok {
				delete(sb.sockets, conn)
				close(conn.send)
			}
		case <-tick:
			update := sb.GetPoint()
		 	
		 	for cnn := range sb.sockets {		
				cnn.send <- update
			}
			
		}
	}
	
}
func (sb *SymbolPush) GetLoopy() {
	ch := time.Tick(1 *time.Second)
		for range ch {
			update := sb.GetPoint()
		 	
		 	for conn := range sb.sockets {		
				conn.send <- update
			}
			
	}
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
	  "AUD/JPY": { "AUD/JPY", "AUD/JPY_feed", 0, 79.013, 87.815, 72.354, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "AUD/USD": { "AUD/USD", "AUD/USD_feed", 0, 0.74521, 0.78337, 0.68265 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "CAD/CHF": { "CAD/CHF", "CAD/CHF_feed", 0, 0.7514, 0.7735, 0.6803, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "CAD/JPY": { "CAD/JPY", "CAD/JPY_feed", 0, 80.757, 88.796, 76.071, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "CHF/JPY": { "CHF/JPY", "CHF/JPY_feed", 0, 107.342, 120.365, 101.195, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "EUR/AUD": { "EUR/AUD", "EUR/AUD_feed", 0, 1.46873, 1.62429, 1.44335 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "EUR/CAD": { "EUR/CAD", "EUR/CAD_feed", 0, 1.44065, 1.61045, 1.41772 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "EUR/GBP": { "EUR/GBP", "EUR/GBP_feed", 0, 0.83713, 0.86255, 0.73119, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "EUR/JPY": { "EUR/JPY", "EUR/JPY_feed", 0, 116.393, 132.285, 109.552, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "EUR/NOK": { "EUR/NOK", "EUR/NOK_feed", 0, 9.37867, 9.74633, 9.15502 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "EUR/SEK": { "EUR/SEK", "EUR/SEK_feed", 0, 9.49165, 9.61074, 9.11289 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "EUR/USD": { "EUR/USD", "EUR/USD_feed", 0, 1.09758, 1.1616, 1.07102, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "GBP/AUD": { "GBP/AUD", "GBP/AUD_feed", 0, 1.75303, 2.09769, 1.7043, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "GBP/CHF": { "GBP/CHF", "GBP/CHF_feed", 0, 1.2927, 1.48345, 1.25059 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "GBP/JPY": { "GBP/JPY", "GBP/JPY_feed", 0, 138.921, 177.356, 128.788 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "GBP/USD": { "GBP/USD", "GBP/USD_feed", 0, 1.31064, 1.50155, 1.27975 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "NZD/JPY": { "NZD/JPY", "NZD/JPY_feed", 0, 74.079, 82.23, 68.889, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "NZD/USD": { "NZD/USD", "NZD/USD_feed", 0, 0.6988, 0.73236, 0.63463, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "USD/BRL": { "USD/BRL", "USD/BRL_feed", 0, 3.25397, 4.16152, 3.18546, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "USD/CAD": { "USD/CAD", "USD/CAD_feed", 0, 1.31252, 1.46896, 1.24602 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/CHF": { "USD/CHF", "USD/CHF_feed", 0, 0.98669, 1.02558, 0.94439 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/CNY": { "USD/CNY", "USD/CNY_feed", 0, 6.6758, 6.7134, 6.4108, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "USD/CZK": { "USD/CZK", "USD/CZK_feed", 0, 24.6089, 25.2263, 23.2712 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/HKD": { "USD/HKD", "USD/HKD_feed", 0, 7.7566, 7.829, 7.7493, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "USD/INR": { "USD/INR", "USD/INR_feed", 0, 67.0958, 68.8808, 65.88 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/JPY": { "USD/JPY", "USD/JPY_feed", 0, 106.033, 121.676, 99.075, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	  "USD/KRW": { "USD/KRW", "USD/KRW_feed", 0, 1134.52, 1244.1, 1122.87 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/MXN": { "USD/MXN", "USD/MXN_feed", 0, 18.5388, 19.5088, 17.0489 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/NOK": { "USD/NOK", "USD/NOK_feed", 0, 8.54192, 8.99242, 7.96395 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/PLN": { "USD/PLN", "USD/PLN_feed", 0, 3.9742, 4.1552, 3.7043, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	  "USD/RUB": { "USD/RUB", "USD/RUB_feed", 0, 64.745, 85.099, 62.791 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/SEK": { "USD/SEK", "USD/SEK_feed", 0, 8.64657, 8.74226, 7.89232 , make(map[*Client]bool), make(chan *Client), make(chan *Client)},
	  "USD/SGD": { "USD/SGD", "USD/SGD_feed", 0, 1.3579, 1.4443, 1.3311, make(map[*Client]bool) , make(chan *Client), make(chan *Client)},
	"ACM": {"ACM", "ACM_feed", 0, 29.33, 35.86, 22.8, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"AKS": {"AKS", "AKS_feed", 0, 3.91, 6.18, 1.64, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"AMBR": {"AMBR", "AMBR_feed", 0, 5.99, 8.55, 3.42, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ARMK": {"ARMK", "ARMK_feed", 0, 32.16, 36.24, 28.09, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BANC": {"BANC", "BANC_feed", 0, 16.95, 22.19, 11.71, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BSBR": {"BSBR", "BSBR_feed", 0, 4.54, 6.13, 2.96, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BAX": {"BAX", "BAX_feed", 0, 40.09, 48.01, 32.18, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BHL": {"BHL", "BHL_feed", 0, 12.97, 13.48, 12.467, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BERY": {"BERY", "BERY_feed", 0, 34.59, 41.4, 27.79, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LND": {"LND", "LND_feed", 0, 3.10, 4.06, 2.15, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BR": {"BR", "BR_feed", 0, 58.47, 68.37, 48.56, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CAT": {"CAT", "CAT_feed", 0, 69.56, 82.75, 56.36, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"IGR": {"IGR", "IGR_feed", 0, 7.40, 8.49, 6.31, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CIM": {"CIM", "CIM_feed", 0, 13.67, 16.44, 10.89, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MJN": {"MJN", "MJN_feed", 0, 79.27, 93, 65.53, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MSCI": {"MSCI", "MSCI_feed", 0, 70.22, 83.56, 56.87, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MUSA": {"MUSA", "MUSA_feed", 0, 63.02, 78.31, 47.73, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NCI": {"NCI", "NCI_feed", 0, 16.39, 19.64, 13.13, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NXRT": {"NXRT", "NXRT_feed", 0, 15.12, 19.89, 10.35, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NTC": {"NTC", "NTC_feed", 0, 13.21, 14.28, 12.14, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NEV": {"NEV", "NEV_feed", 0, 15.71, 16.96, 14.45, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PCI": {"PCI", "PCI_feed", 0, 18.29, 19.99, 16.596, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PHK": {"PHK", "PHK_feed", 0, 8.44, 10.01, 6.871, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"RCS": {"RCS", "RCS_feed", 0, 8.25, 10.48, 6.01, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PPX": {"PPX", "PPX_feed", 0, 26.18, 27.74, 24.63, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PLD": {"PLD", "PLD_feed", 0, 44.23, 53.2, 35.25, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PRH": {"PRH", "PRH_feed", 0, 25.83, 27.18, 24.48, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"TLK": {"TLK", "TLK_feed", 0, 49.59, 65.1, 34.09, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PSA": {"PSA", "PSA_feed", 0, 26.59, 28.19, 25, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CTY": {"CTY", "CTY_feed", 0, 24.51, 25.76, 23.25, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CTAA": {"CTAA", "CTAA_feed", 0, 25.98, 27.7, 24.25, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"RFT": {"RFT", "RFT_feed", 0, 18.63, 23.5, 13.75, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SAP": {"SAP", "SAP_feed", 0, 73.95, 85.33, 62.57, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"AOS": {"AOS", "AOS_feed", 0, 71.91, 93.72, 50.09, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SWK": {"SWK", "SWK_feed", 0, 105.57, 122.42, 88.72, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"EDF": {"EDF", "EDF_feed", 0, 12.57, 15.19, 9.95, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SRI": {"SRI", "SRI_feed", 0, 13.68, 17.18, 10.18, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BLD": {"BLD", "BLD_feed", 0, 30.84, 38.65, 23.02, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"TTC": {"TTC", "TTC_feed", 0, 78.60, 92.49, 64.71, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"TDG": {"TDG", "TDG_feed", 0, 229.03, 277.3, 180.76, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"UTX": {"UTX", "UTX_feed", 0, 95.64, 107.89, 83.39, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"WCN": {"WCN", "WCN_feed", 0, 58.83, 74.44, 43.219, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"WFC": {"WFC", "WFC_feed", 0, 24.79, 26.15, 23.42, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"OILD": {"OILD", "OILD_feed", 0, 24.46, 26.7, 22.23, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"AEIS": {"AEIS", "AEIS_feed", 0, 30.81, 40.51, 21.12, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"AMD": {"AMD", "AMD_feed", 0, 4.29, 6.98, 1.61, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ALGN": {"ALGN", "ALGN_feed", 0, 68.81, 85.61, 52.01, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"APEI": {"APEI", "APEI_feed", 0, 22.05, 30.29, 13.8, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ASML": {"ASML", "ASML_feed", 0, 92.97, 108.77, 77.173, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ASTE": {"ASTE", "ASTE_feed", 0, 45.47, 60.18, 30.76, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ALOT": {"ALOT", "ALOT_feed", 0, 13.59, 16, 11.18, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ATRI": {"ATRI", "ATRI_feed", 0, 407.95, 472.4, 343.5, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ADRE": {"ADRE", "ADRE_feed", 0, 26.30, 33.04, 19.552, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BMCH": {"BMCH", "BMCH_feed", 0, 16.29, 20.44, 12.14, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"BRKS": {"BRKS", "BRKS_feed", 0, 10.62, 12.9, 8.33, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CA": {"CA", "CA_feed", 0, 29.92, 34.68, 25.156, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CPLA": {"CPLA", "CPLA_feed", 0, 49.83, 60.59, 39.06, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CARO": {"CARO", "CARO_feed", 0, 17.25, 20, 14.49, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CASY": {"CASY", "CASY_feed", 0, 115.81, 135.92, 95.71, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CBNK": {"CBNK", "CBNK_feed", 0, 17.45, 18.89, 16, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CSCO": {"CSCO", "CSCO_feed", 0, 26.67, 30.88, 22.46, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CVLT": {"CVLT", "CVLT_feed", 0, 40.43, 51.45, 29.41, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CYBE": {"CYBE", "CYBE_feed", 0, 12.07, 19.34, 4.8, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CYNO": {"CYNO", "CYNO_feed", 0, 41.63, 54.25, 29, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CYRN": {"CYRN", "CYRN_feed", 0, 1.73, 2.17, 1.3, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"DLTR": {"DLTR", "DLTR_feed", 0, 78.48, 96.65, 60.31, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"DORM": {"DORM", "DORM_feed", 0, 50.51, 60.84, 40.17, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"EBAY": {"EBAY", "EBAY_feed", 0, 26.46, 31.4, 21.515, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FNTCW": {"FNTCW", "FNTCW_feed", 0, 0.68, 1.15, 0.22, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FFBC": {"FFBC", "FFBC_feed", 0, 17.54, 21.31, 13.76, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"THFF": {"THFF", "THFF_feed", 0, 34.99, 38.83, 31.15, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SKYY": {"SKYY", "SKYY_feed", 0, 28.33, 32.16, 24.5, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FEX": {"FEX", "FEX_feed", 0, 33.58, 47.05, 20.11, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FTC": {"FTC", "FTC_feed", 0, 39.80, 51.57, 28.02, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FTA": {"FTA", "FTA_feed", 0, 30.75, 42.21, 19.285, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FAB": {"FAB", "FAB_feed", 0, 39.15, 45.081, 33.22, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"TDIV": {"TDIV", "TDIV_feed", 0, 22.00, 28.59, 15.4, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"QTEC": {"QTEC", "QTEC_feed", 0, 40.55, 47.25, 33.85, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FYX": {"FYX", "FYX_feed", 0, 40.61, 48.83, 32.38, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FIVE": {"FIVE", "FIVE_feed", 0, 39.41, 51.87, 26.95, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FELE": {"FELE", "FELE_feed", 0, 31.32, 38.89, 23.75, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"GRMN": {"GRMN", "GRMN_feed", 0, 38.66, 46.39, 30.93, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"GENC": {"GENC", "GENC_feed", 0, 13.29, 17.68, 8.91, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"GNTX": {"GNTX", "GNTX_feed", 0, 15.37, 17.8, 12.93, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SOCL": {"SOCL", "SOCL_feed", 0, 18.56, 21.92, 15.2, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"GDEN": {"GDEN", "GDEN_feed", 0, 10.71, 13.52, 7.89, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"GBDC": {"GBDC", "GBDC_feed", 0, 16.77, 18.75, 14.8, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"IESC": {"IESC", "IESC_feed", 0, 11.29, 16.3, 6.28, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"IPAR": {"IPAR", "IPAR_feed", 0, 26.73, 33.1, 20.368, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SOXX": {"SOXX", "SOXX_feed", 0, 88.96, 105.05, 72.863, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"JKHY": {"JKHY", "JKHY_feed", 0, 76.61, 89.37, 63.84, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"JSMD": {"JSMD", "JSMD_feed", 0, 27.25, 29.48, 25.02, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"KEQU": {"KEQU", "KEQU_feed", 0, 17.90, 19.99, 15.815, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"KLAC": {"KLAC", "KLAC_feed", 0, 60.97, 76.99, 44.95, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LKFN": {"LKFN", "LKFN_feed", 0, 45.19, 51.37, 39.01, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LRCX": {"LRCX", "LRCX_feed", 0, 76.40, 91.6, 61.2, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LXRX": {"LXRX", "LXRX_feed", 0, 11.93, 16.22, 7.65, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LGIH": {"LGIH", "LGIH_feed", 0, 27.16, 36.2, 18.12, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LSXMB": {"LSXMB", "LSXMB_feed", 0, 32.34, 36.08, 28.6, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LSXMA": {"LSXMA", "LSXMA_feed", 0, 31.65, 35.3, 28, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LLTC": {"LLTC", "LLTC_feed", 0, 49.45, 62.49, 36.41, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"LITE": {"LITE", "LITE_feed", 0, 22.34, 30.72, 13.97, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MGIC": {"MGIC", "MGIC_feed", 0, 6.25, 7.37, 5.13, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MBFI": {"MBFI", "MBFI_feed", 0, 33.41, 38.85, 27.98, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MGPI": {"MGPI", "MGPI_feed", 0, 27.59, 42.32, 12.85, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MCHP": {"MCHP", "MCHP_feed", 0, 47.56, 57.34, 37.77, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MKSI": {"MKSI", "MKSI_feed", 0, 38.23, 47.47, 29, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"MPWR": {"MPWR", "MPWR_feed", 0, 57.27, 73.79, 40.752, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NANO": {"NANO", "NANO_feed", 0, 17.29, 22.66, 11.91, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NGHCZ": {"NGHCZ", "NGHCZ_feed", 0, 24.36, 25.97, 22.75, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NATL": {"NATL", "NATL_feed", 0, 26.79, 32.43, 21.148, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NTES": {"NTES", "NTES_feed", 0, 153.65, 204.49, 102.8, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NVEE": {"NVEE", "NVEE_feed", 0, 23.75, 32.5, 15, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"NVDA": {"NVDA", "NVDA_feed", 0, 37.86, 56.63, 19.09, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"OCLR": {"OCLR", "OCLR_feed", 0, 4.04, 6.01, 2.08, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ORLY": {"ORLY", "ORLY_feed", 0, 253.44, 281.77, 225.12, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PDFS": {"PDFS", "PDFS_feed", 0, 12.40, 16.11, 8.7, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PEBO": {"PEBO", "PEBO_feed", 0, 19.36, 22.37, 16.34, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PEBK": {"PEBK", "PEBK_feed", 0, 18.52, 21, 16.03, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PRFT": {"PRFT", "PRFT_feed", 0, 18.55, 22.21, 14.9, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"POWI": {"POWI", "POWI_feed", 0, 45.93, 55.82, 36.04, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PSCI": {"PSCI", "PSCI_feed", 0, 42.84, 48.55, 37.134, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PSCT": {"PSCT", "PSCT_feed", 0, 49.02, 58.63, 39.41, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PSEC": {"PSEC", "PSEC_feed", 0, 6.82, 8.43, 5.21, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"PROV": {"PROV", "PROV_feed", 0, 17.57, 19.63, 15.51, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"QCRH": {"QCRH", "QCRH_feed", 0, 23.58, 29.1, 18.05, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"RAVN": {"RAVN", "RAVN_feed", 0, 16.66, 20.44, 12.88, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"ROST": {"ROST", "ROST_feed", 0, 52.72, 61.98, 43.47, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SMTC": {"SMTC", "SMTC_feed", 0, 19.67, 25.31, 14.04, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SIRI": {"SIRI", "SIRI_feed", 0, 3.82, 4.35, 3.29, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"JSM": {"JSM", "JSM_feed", 0, 17.20, 21.86, 12.54, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"OKSB": {"OKSB", "OKSB_feed", 0, 16.67, 19.34, 14, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SPAR": {"SPAR", "SPAR_feed", 0, 5.30, 7.99, 2.61, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"TXN": {"TXN", "TXN_feed", 0, 57.45, 71.42, 43.49, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"SPNC": {"SPNC", "SPNC_feed", 0, 16.30, 21.95, 10.65, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"UBNT": {"UBNT", "UBNT_feed", 0, 35.00, 44.26, 25.75, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"UFPI": {"UFPI", "UFPI_feed", 0, 82.28, 107.46, 57.11, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"VTWV": {"VTWV", "VTWV_feed", 0, 78.29, 88.445, 68.129, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"VASC": {"VASC", "VASC_feed", 0, 35.35, 46.36, 24.34, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CFO": {"CFO", "CFO_feed", 0, 35.26, 38.82, 31.69, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CFA": {"CFA", "CFA_feed", 0, 35.29, 38.83, 31.75, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"CDL": {"CDL", "CDL_feed", 0, 35.18, 38.801, 31.55, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
	"FLAG": {"FLAG", "FLAG_feed", 0, 30.30, 33.861, 26.73, make(map[*Client]bool), make(chan *Client), make(chan *Client) },
}
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	addClient    chan *Client
	removeClient chan *Client
}

var hub = Hub{
	broadcast:    make(chan []byte),
	addClient:    make(chan *Client),
	removeClient: make(chan *Client),
	clients:      make(map[*Client]bool),
}

func (hub *Hub) start() {
	for {
		select {
		case conn := <-hub.addClient:
			hub.clients[conn] = true
		case conn := <-hub.removeClient:
			if _, ok := hub.clients[conn]; ok {
				delete(hub.clients, conn)
				close(conn.send)
			}
		}
	}
}

func (c *Client) write() {
	defer func() {
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.ws.WriteJSON(message)

		}
	}
}

func (c *Client) read() {
	defer func() {
		hub.removeClient <- c
		c.ws.Close()
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		str :=  string(message[:])
		seriesPush := SymStreams[str]
		seriesPush.addClient <- c
		fmt.Println(str)
		c.subs = append(c.subs, str)
		if err != nil {
			hub.removeClient <- c
			for i := 0; i < len(c.subs); i++ {
			  SymStreams[c.subs[i]].removeClient <- c
			}
			c.ws.Close()
			break
		}
        fmt.Println(hub.clients)
        fmt.Println(c.subs)
	}
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		http.NotFound(res, req)
		return
	}

	client := &Client{
		ws:   conn,
		send: make(chan SeriesPoint),
		subs: []string{},
	}

	hub.addClient <- client

	go client.write()
	go client.read()
}

func main() {
	go hub.start()
	for v  := range SymStreams{
		w := SymStreams[v]
		go w.start()
	}
	http.HandleFunc("/", hello)
	
	onOpenShift := false
    if ( onOpenShift) {
		bind := fmt.Sprintf("%s:%s", os.Getenv("OPENSHIFT_GO_IP"), os.Getenv("OPENSHIFT_GO_PORT"))
		fmt.Printf("listening on %s...", bind)
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			panic(err)
		}
	}
	
		


	http.HandleFunc("/indi", func(w http.ResponseWriter, r *http.Request){
		http.ServeFile(w, r, "index.html")
    })
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request){
		wsPage(w, r)
	
    })
	http.ListenAndServe(":8443",nil)
	
		
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "hello, world from %s", runtime.Version())
}