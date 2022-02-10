package dvfapi

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type OrderBookBranch struct {
	Bids          BookBranch
	Asks          BookBranch
	LastUpdatedId decimal.Decimal
	SnapShoted    bool
	Cancel        *context.CancelFunc
	reCh          chan error
	lastRefresh   lastRefreshBranch
}

type lastRefreshBranch struct {
	mux  sync.RWMutex
	time time.Time
}

type BookBranch struct {
	mux  sync.RWMutex
	Book [][]string
}

func (o *OrderBookBranch) IfCanRefresh() bool {
	o.lastRefresh.mux.Lock()
	defer o.lastRefresh.mux.Unlock()
	now := time.Now()
	if now.After(o.lastRefresh.time.Add(time.Second * 3)) {
		o.lastRefresh.time = now
		return true
	}
	return false
}

func (o *OrderBookBranch) UpdateNewComing(message *map[string]interface{}) {
	var wg sync.WaitGroup
	data := (*message)["data"].([]interface{})
	for _, item := range data {
		if book, ok := item.(map[string]interface{}); ok {
			if bids, ok := book["bids"].([]interface{}); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, item := range bids {
						if levelData, ok := item.([]interface{}); ok {
							price, okPrice := levelData[0].(string)
							size, okSize := levelData[1].(string)
							if !okPrice || !okSize {
								continue
							}
							decPrice, _ := decimal.NewFromString(price)
							decSize, _ := decimal.NewFromString(size)
							o.DealWithBidPriceLevel(decPrice, decSize)
						}
					}
				}()
			}
			if asks, ok := book["asks"].([]interface{}); ok {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, item := range asks {
						if levelData, ok := item.([]interface{}); ok {
							price, okPrice := levelData[0].(string)
							size, okSize := levelData[1].(string)
							if !okPrice || !okSize {
								continue
							}
							decPrice, _ := decimal.NewFromString(price)
							decSize, _ := decimal.NewFromString(size)
							o.DealWithAskPriceLevel(decPrice, decSize)
						}
					}
				}()
			}
			wg.Wait()
		}
	}
}

func (o *OrderBookBranch) DealWithBidPriceLevel(price, qty decimal.Decimal) {
	o.Bids.mux.Lock()
	defer o.Bids.mux.Unlock()
	l := len(o.Bids.Book)
	if l == 0 {
		o.Bids.Book = append(o.Bids.Book, []string{price.String(), qty.String()})
		return
	}
	for level, item := range o.Bids.Book {
		bookPrice, _ := decimal.NewFromString(item[0])
		switch {
		case price.GreaterThan(bookPrice):
			// insert level
			if qty.IsZero() {
				// ignore
				return
			}
			o.Bids.Book = append(o.Bids.Book, []string{})
			copy(o.Bids.Book[level+1:], o.Bids.Book[level:])
			o.Bids.Book[level] = []string{price.String(), qty.String()}
			return
		case price.LessThan(bookPrice):
			if level == l-1 {
				// insert last level
				if qty.IsZero() {
					// ignore
					return
				}
				o.Bids.Book = append(o.Bids.Book, []string{price.String(), qty.String()})
				return
			}
			continue
		case price.Equal(bookPrice):
			if qty.IsZero() {
				// delete level
				o.Bids.Book = append(o.Bids.Book[:level], o.Bids.Book[level+1:]...)
				return
			}
			o.Bids.Book[level][1] = qty.String()
			return
		}
	}
}

func (o *OrderBookBranch) DealWithAskPriceLevel(price, qty decimal.Decimal) {
	o.Asks.mux.Lock()
	defer o.Asks.mux.Unlock()
	l := len(o.Asks.Book)
	if l == 0 {
		o.Asks.Book = append(o.Asks.Book, []string{price.String(), qty.String()})
		return
	}
	for level, item := range o.Asks.Book {
		bookPrice, _ := decimal.NewFromString(item[0])
		switch {
		case price.LessThan(bookPrice):
			// insert level
			if qty.IsZero() {
				// ignore
				return
			}
			o.Asks.Book = append(o.Asks.Book, []string{})
			copy(o.Asks.Book[level+1:], o.Asks.Book[level:])
			o.Asks.Book[level] = []string{price.String(), qty.String()}
			return
		case price.GreaterThan(bookPrice):
			if level == l-1 {
				// insert last level
				if qty.IsZero() {
					// ignore
					return
				}
				o.Asks.Book = append(o.Asks.Book, []string{price.String(), qty.String()})
				return
			}
			continue
		case price.Equal(bookPrice):
			if qty.IsZero() {
				// delete level
				o.Asks.Book = append(o.Asks.Book[:level], o.Asks.Book[level+1:]...)
				return
			}
			o.Asks.Book[level][1] = qty.String()
			return
		}
	}
}

func (o *OrderBookBranch) RefreshLocalOrderBook(err error) error {
	if o.IfCanRefresh() {
		if len(o.reCh) == cap(o.reCh) {
			return errors.New("refresh channel is full, please check it up")
		}
		o.reCh <- err
	}
	return nil
}

func (o *OrderBookBranch) Close() {
	(*o.Cancel)()
	o.SnapShoted = false
	o.Bids.mux.Lock()
	o.Bids.Book = [][]string{}
	o.Bids.mux.Unlock()
	o.Asks.mux.Lock()
	o.Asks.Book = [][]string{}
	o.Asks.mux.Unlock()
}

// return bids, ready or not
func (o *OrderBookBranch) GetBids() ([][]string, bool) {
	o.Bids.mux.RLock()
	defer o.Bids.mux.RUnlock()
	if !o.SnapShoted {
		return [][]string{}, false
	}
	if len(o.Bids.Book) == 0 {
		if o.IfCanRefresh() {
			o.reCh <- errors.New("re cause len bid is zero")
		}
		return [][]string{}, false
	}
	book := o.Bids.Book
	return book, true
}

func (o *OrderBookBranch) GetBidsEnoughForValue(value decimal.Decimal) ([][]string, bool) {
	o.Bids.mux.RLock()
	defer o.Bids.mux.RUnlock()
	if len(o.Bids.Book) == 0 || !o.SnapShoted {
		return [][]string{}, false
	}
	var loc int
	var sumValue decimal.Decimal
	for level, data := range o.Bids.Book {
		if len(data) != 2 {
			return [][]string{}, false
		}
		price, _ := decimal.NewFromString(data[0])
		size, _ := decimal.NewFromString(data[1])
		sumValue = sumValue.Add(price.Mul(size))
		if sumValue.GreaterThan(value) {
			loc = level
			break
		}
	}
	book := o.Bids.Book[:loc+1]
	return book, true
}

// return asks, ready or not
func (o *OrderBookBranch) GetAsks() ([][]string, bool) {
	o.Asks.mux.RLock()
	defer o.Asks.mux.RUnlock()
	if !o.SnapShoted {
		return [][]string{}, false
	}
	if len(o.Asks.Book) == 0 {
		if o.IfCanRefresh() {
			o.reCh <- errors.New("re cause len ask is zero")
		}
		return [][]string{}, false
	}
	book := o.Asks.Book
	return book, true
}

func (o *OrderBookBranch) GetAsksEnoughForValue(value decimal.Decimal) ([][]string, bool) {
	o.Asks.mux.RLock()
	defer o.Asks.mux.RUnlock()
	if len(o.Asks.Book) == 0 || !o.SnapShoted {
		return [][]string{}, false
	}
	var loc int
	var sumValue decimal.Decimal
	for level, data := range o.Asks.Book {
		if len(data) != 2 {
			return [][]string{}, false
		}
		price, _ := decimal.NewFromString(data[0])
		size, _ := decimal.NewFromString(data[1])
		sumValue = sumValue.Add(price.Mul(size))
		if sumValue.GreaterThan(value) {
			loc = level
			break
		}
	}
	book := o.Asks.Book[:loc+1]
	return book, true
}

// symbol example: ETH:USDT
func LocalOrderBook(symbol string, logger *log.Logger) *OrderBookBranch {
	var o OrderBookBranch
	ctx, cancel := context.WithCancel(context.Background())
	o.Cancel = &cancel
	bookticker := make(chan map[string]interface{}, 50)
	refreshCh := make(chan error, 5)
	o.reCh = make(chan error, 5)
	symbol = strings.ToUpper(symbol)
	url := SocketEndPointHub(false)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := DVFOrderBookSocket(ctx, url, symbol, "orderbook", logger, &bookticker, &refreshCh); err == nil {
					return
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := o.MaintainOrderBook(ctx, symbol, &bookticker)
				if err == nil {
					return
				}
				logger.Warningf("refreshing %s local orderbook cause: %s", symbol, err.Error())
				refreshCh <- errors.New("refreshing from maintain orderbook")
			}
		}
	}()
	return &o
}

func (o *OrderBookBranch) MaintainOrderBook(
	ctx context.Context,
	symbol string,
	bookticker *chan map[string]interface{},
) error {
	//var storage []map[string]interface{}
	o.SnapShoted = false
	o.LastUpdatedId = decimal.NewFromInt(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-o.reCh:
			return err
		default:
			message := <-(*bookticker)
			if len(message) != 0 {
				// for initial orderbook
				if snapshot, ok := message["snapshot"]; ok {
					o.InitialOrderBook(&snapshot)
					continue
				}
				if update, ok := message["update"]; ok {
					o.SpotUpdateJudge(&update)
					continue
				}
			}
		}
	}
}

func (o *OrderBookBranch) SpotUpdateJudge(res *interface{}) {
	data := (*res).([]interface{})
	price := decimal.NewFromFloat(data[0].(float64))
	qty := decimal.NewFromFloat(data[2].(float64))
	count := data[1].(float64)
	realQty := qty.Abs()
	if count == 0 {
		realQty = decimal.Zero
	}
	if qty.IsPositive() {
		o.DealWithBidPriceLevel(price, realQty)
	} else {
		o.DealWithAskPriceLevel(price, realQty)
	}
}

func (o *OrderBookBranch) InitialOrderBook(res *interface{}) {
	//var wg sync.WaitGroup
	set := (*res).([]interface{})
	for _, item := range set {
		data := item.([]interface{})
		price := decimal.NewFromFloat(data[0].(float64))
		qty := decimal.NewFromFloat(data[2].(float64))
		if qty.IsPositive() {
			o.DealWithBidPriceLevel(price, qty)
		} else {
			o.DealWithAskPriceLevel(price, qty.Abs())
		}
	}
	o.SnapShoted = true
}

type DVFWebsocket struct {
	Channel       string
	OnErr         bool
	Logger        *log.Logger
	Conn          *websocket.Conn
	LastUpdatedId decimal.Decimal
	ChannelID     float64
}

type DVFSubscribeMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol"`
}

func (w *DVFWebsocket) OutDVFErr() map[string]interface{} {
	w.OnErr = true
	m := make(map[string]interface{})
	return m
}

func DecodingMap(message *[]byte, logger *log.Logger) (res interface{}, err error) {
	if *message == nil {
		err = errors.New("the incoming message is nil")
		return nil, err
	}
	err = json.Unmarshal(*message, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DVFOrderBookSocket(
	ctx context.Context,
	url, symbol, channel string,
	logger *log.Logger,
	mainCh *chan map[string]interface{},
	refreshCh *chan error,
) error {
	var w DVFWebsocket
	var duration time.Duration = 30
	w.Logger = logger
	w.OnErr = false
	w.ChannelID = 0
	innerErr := make(chan error, 1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	logger.Infof("DVF %s orderBook socket connected.\n", symbol)
	w.Conn = conn
	defer conn.Close()
	send := GetDVFSubscribeMessage(channel, symbol)
	if err := w.Conn.WriteMessage(websocket.TextMessage, send); err != nil {
		return err
	}
	if err := w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
		return err
	}
	read := time.NewTicker(time.Millisecond * 50)
	defer read.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-*refreshCh:
			innerErr <- errors.New("restart")
			return err
		case <-read.C:
			if conn == nil {
				d := w.OutDVFErr()
				*mainCh <- d
				message := "DVF reconnect..."
				logger.Infoln(message)
				innerErr <- errors.New("restart")
				return errors.New(message)
			}
			_, buf, err := conn.ReadMessage()
			if err != nil {
				d := w.OutDVFErr()
				*mainCh <- d
				message := "DVF reconnect..."
				logger.Infoln(message)
				innerErr <- errors.New("restart")
				return errors.New(message)
			}
			res, err1 := DecodingMap(&buf, logger)
			if err1 != nil {
				d := w.OutDVFErr()
				*mainCh <- d
				message := "DVF reconnect..."
				logger.Infoln(message, err1)
				innerErr <- errors.New("restart")
				return err1
			}
			err2 := w.HandleDVFSocketData(&res, mainCh)
			if err2 != nil {
				d := w.OutDVFErr()
				*mainCh <- d
				message := "DVF reconnect..."
				logger.Infoln(message, err2)
				innerErr <- errors.New("restart")
				return err2
			}
			if err := w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
				return err
			}
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (w *DVFWebsocket) HandleDVFSocketData(res *interface{}, mainCh *chan map[string]interface{}) error {
	if dataSet, ok := (*res).([]interface{}); ok {
		if w.ChannelID == 0 {
			// initial orderbook
			if id, ok := dataSet[0].(float64); ok {
				w.ChannelID = id
			}
			if book, ok := dataSet[1].([]interface{}); ok {
				data := make(map[string]interface{})
				data["snapshot"] = book
				*mainCh <- data
				return nil
			} else {
				return errors.New("fail to initial orderbook")
			}
		}
		// update part
		if id, ok := dataSet[0].(float64); ok {
			if id != w.ChannelID {
				return errors.New("wrong channel id return")
			}
			if book, ok := dataSet[1].([]interface{}); ok {
				data := make(map[string]interface{})
				data["update"] = book
				*mainCh <- data
				return nil
			} else {
				return errors.New("fail to update orderbook")
			}
		}

	}
	return nil
}

func GetDVFSubscribeMessage(channel, symbol string) (message []byte) {
	switch channel {
	case "orderbook":
		sub := DVFSubscribeMessage{
			Event:   "subscribe",
			Channel: "book",
			Symbol:  symbol,
		}
		by, err := json.Marshal(sub)
		if err != nil {
			return nil
		}
		message = by
	}
	return message
}
