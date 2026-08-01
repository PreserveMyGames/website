package fx

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PreserveMyGames/store/internal/constants"
)

type Rates struct {
	Base  string
	AsOf  time.Time
	Table map[string]float64
}

type Cache struct {
	mu     sync.RWMutex
	rates  Rates
	client *http.Client
}

func New() *Cache {
	c := &Cache{
		client: &http.Client{Timeout: 10 * time.Second},
		rates: Rates{
			Base:  constants.DefaultBaseCurrency,
			AsOf:  time.Time{},
			Table: map[string]float64{constants.DefaultBaseCurrency: 1},
		},
	}
	return c
}

func (c *Cache) Start() {
	c.Refresh()
	go func() {
		t := time.NewTicker(constants.FXRefreshInterval)
		defer t.Stop()
		for range t.C {
			c.Refresh()
		}
	}()
}

func (c *Cache) Refresh() {
	req, err := http.NewRequest(http.MethodGet, "https://api.frankfurter.app/latest?from=USD", nil)
	if err != nil {
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	var payload struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return
	}
	table := map[string]float64{strings.ToUpper(payload.Base): 1}
	for k, v := range payload.Rates {
		table[strings.ToUpper(k)] = v
	}
	c.mu.Lock()
	c.rates = Rates{Base: strings.ToUpper(payload.Base), AsOf: time.Now().UTC(), Table: table}
	c.mu.Unlock()
}

func (c *Cache) Convert(cents int64, from, to string) int64 {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	if from == to || cents == 0 {
		return cents
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	fromRate, okFrom := c.rates.Table[from]
	toRate, okTo := c.rates.Table[to]
	if !okFrom || !okTo || fromRate == 0 {
		return cents
	}
	// Convert via USD base table from Frankfurter.
	usd := float64(cents) / fromRate
	out := usd * toRate
	return int64(math.Round(out))
}

func (c *Cache) FormatConverted(cents int64, from, to string) string {
	converted := c.Convert(cents, from, to)
	return fmt.Sprintf("%s %.2f", strings.ToUpper(to), float64(converted)/100)
}

func (c *Cache) Supported() []string {
	return []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF", "SEK", "NOK", "DKK", "PLN", "CZK"}
}
