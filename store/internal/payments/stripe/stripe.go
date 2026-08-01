package stripepay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/constants"
)

type Client struct {
	secretKey     string
	webhookSecret string
}

func New(secretKey, webhookSecret string) *Client {
	stripe.Key = secretKey
	return &Client{secretKey: secretKey, webhookSecret: webhookSecret}
}

func (c *Client) CreateCheckout(order catalog.Order, product catalog.Product, successURL, cancelURL string) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(stringsLower(product.Currency)),
					UnitAmount: stripe.Int64(product.PriceCents),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(product.Title),
					},
				},
			},
		},
		ClientReferenceID: stripe.String(order.ID),
		Metadata: map[string]string{
			"order_id": order.ID,
		},
	}
	s, err := session.New(params)
	if err != nil {
		return "", "", err
	}
	return s.ID, s.URL, nil
}

func (c *Client) HandleWebhook(payload []byte, sigHeader string) (orderID string, err error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, c.webhookSecret)
	if err != nil {
		return "", err
	}
	if event.Type != "checkout.session.completed" {
		return "", nil
	}
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		return "", err
	}
	if sess.ClientReferenceID != "" {
		return sess.ClientReferenceID, nil
	}
	if sess.Metadata != nil {
		return sess.Metadata["order_id"], nil
	}
	return "", fmt.Errorf("missing order id in stripe session")
}

func ReadBody(r *http.Request, max int64) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(io.LimitReader(r.Body, max))
}

func ProviderName() string {
	return constants.PaymentStripe
}

func stringsLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
