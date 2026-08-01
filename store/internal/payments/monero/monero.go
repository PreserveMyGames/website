package monero

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/constants"
)

type Wallet interface {
	CreateAddress(label string) (address string, index int, err error)
	GetTransfers(accountIndex int) ([]Transfer, error)
}

type Transfer struct {
	Amount        uint64
	Confirmations uint64
	TxID          string
	SubaddrIndex  int
	Unlocked      bool
}

type RPCClient struct {
	url    string
	user   string
	pass   string
	client *http.Client
	mu     sync.Mutex
	nextID int
}

func NewRPC(url, user, pass string) *RPCClient {
	return &RPCClient{
		url:    url,
		user:   user,
		pass:   pass,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *RPCClient) call(method string, params any, result any) error {
	c.mu.Lock()
	c.nextID++
	id := strconv.Itoa(c.nextID)
	c.mu.Unlock()

	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.url+"/json_rpc", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.user != "" {
		req.SetBasicAuth(c.user, c.pass)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var out rpcResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if out.Error != nil {
		return fmt.Errorf("monero rpc %s: %s", method, out.Error.Message)
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(out.Result, result)
}

func (c *RPCClient) CreateAddress(label string) (string, int, error) {
	var res struct {
		Address      string `json:"address"`
		AddressIndex int    `json:"address_index"`
	}
	err := c.call("create_address", map[string]any{
		"account_index": 0,
		"label":         label,
	}, &res)
	return res.Address, res.AddressIndex, err
}

func (c *RPCClient) GetTransfers(accountIndex int) ([]Transfer, error) {
	var res struct {
		In []struct {
			Amount        uint64 `json:"amount"`
			Confirmations uint64 `json:"confirmations"`
			TxID          string `json:"txid"`
			SubaddrIndex  struct {
				Major uint64 `json:"major"`
				Minor uint64 `json:"minor"`
			} `json:"subaddr_index"`
			Unlocked bool `json:"unlocked"`
		} `json:"in"`
	}
	err := c.call("get_transfers", map[string]any{
		"in":            true,
		"account_index": accountIndex,
	}, &res)
	if err != nil {
		return nil, err
	}
	out := make([]Transfer, 0, len(res.In))
	for _, t := range res.In {
		out = append(out, Transfer{
			Amount:        t.Amount,
			Confirmations: t.Confirmations,
			TxID:          t.TxID,
			SubaddrIndex:  int(t.SubaddrIndex.Minor),
			Unlocked:      t.Unlocked,
		})
	}
	return out, nil
}

type Quoter struct {
	client *http.Client
}

func NewQuoter() *Quoter {
	return &Quoter{client: &http.Client{Timeout: 10 * time.Second}}
}

// QuoteAtomic converts fiat cents to atomic XMR units (1 XMR = 1e12 atomic).
func (q *Quoter) QuoteAtomic(cents int64, currency string) (int64, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.coingecko.com/api/v3/simple/price?ids=monero&vs_currencies="+currency, nil)
	if err != nil {
		return 0, err
	}
	resp, err := q.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var payload map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	price, ok := payload["monero"][currency]
	if !ok || price <= 0 {
		return 0, fmt.Errorf("missing monero price for %s", currency)
	}
	fiat := float64(cents) / 100
	xmr := fiat / price
	atomic := int64(math.Round(xmr * 1e12))
	if atomic < 1 {
		atomic = 1
	}
	return atomic, nil
}

func FormatXMR(atomic int64) string {
	return fmt.Sprintf("%.12f", float64(atomic)/1e12)
}

type Poller struct {
	store            *catalog.Store
	wallet           Wallet
	minConfirmations int
	stop             chan struct{}
}

func NewPoller(store *catalog.Store, wallet Wallet, minConfirmations int) *Poller {
	return &Poller{
		store:            store,
		wallet:           wallet,
		minConfirmations: minConfirmations,
		stop:             make(chan struct{}),
	}
}

func (p *Poller) Start() {
	go func() {
		t := time.NewTicker(constants.MoneroPollInterval)
		defer t.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-t.C:
				_ = p.Tick()
			}
		}
	}()
}

func (p *Poller) Stop() {
	close(p.stop)
}

func (p *Poller) Tick() error {
	_ = p.store.ReleaseExpiredReservations()
	orders, err := p.store.ListAwaitingMonero()
	if err != nil {
		return err
	}
	transfers, err := p.wallet.GetTransfers(0)
	if err != nil {
		return err
	}
	byIndex := map[int][]Transfer{}
	for _, t := range transfers {
		byIndex[t.SubaddrIndex] = append(byIndex[t.SubaddrIndex], t)
	}
	for _, o := range orders {
		list := byIndex[o.MoneroAddressIndex]
		var total uint64
		var best Transfer
		for _, t := range list {
			total += t.Amount
			if t.Confirmations >= best.Confirmations {
				best = t
			}
		}
		if total == 0 {
			continue
		}
		_ = p.store.UpdateMoneroProgress(o.ID, best.TxID, int(best.Confirmations))
		// Allow 1% underpayment tolerance for rounding.
		needed := uint64(o.MoneroAmountAtomic)
		minAccept := needed - needed/100
		if total >= minAccept && int(best.Confirmations) >= p.minConfirmations {
			_ = p.store.MarkOrderPaid(o.ID, constants.PaymentMonero, best.TxID)
		}
	}
	return nil
}
