package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"
)

// IyzicoClient, Iyzico Checkout Form API'si için minimum client
// Sandbox ve production base URL'leri destekler
type IyzicoClient struct {
	apiKey    string
	secretKey string
	baseURL   string
	logger    *zap.Logger
	http      *http.Client
}

func NewIyzicoClient(logger *zap.Logger) *IyzicoClient {
	base := os.Getenv("IYZICO_BASE_URL")
	if base == "" {
		base = "https://api.iyzipay.com"
	}
	return &IyzicoClient{
		apiKey:    os.Getenv("IYZICO_API_KEY"),
		secretKey: os.Getenv("IYZICO_SECRET_KEY"),
		baseURL:   base,
		logger:    logger,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *IyzicoClient) IsConfigured() bool {
	return c.apiKey != "" && c.secretKey != ""
}

// CheckoutFormItem, Iyzico'nun sepete eklenen kalem formatı
type CheckoutFormItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category1 string `json:"category1"` // "Service"
	Category2 string `json:"category2,omitempty"`
	ItemType  string `json:"itemType"` // "VIRTUAL" | "PHYSICAL"
	Price     string `json:"price"`    // "4999.00"
}

type CheckoutFormBuyer struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Surname             string `json:"surname"`
	Email               string `json:"email"`
	IdentityNumber      string `json:"identityNumber"`
	RegistrationAddress string `json:"registrationAddress"`
	City                string `json:"city"`
	Country             string `json:"country"`
	ZipCode             string `json:"zipCode"`
	IP                  string `json:"ip"`
}

type CheckoutFormAddress struct {
	Address     string `json:"address"`
	ZipCode     string `json:"zipCode"`
	ContactName string `json:"contactName"`
	City        string `json:"city"`
	Country     string `json:"country"`
}

type CheckoutFormReq struct {
	Locale             string             `json:"locale"`             // "tr" | "en"
	ConversationID     string             `json:"conversationId"`
	Price              string             `json:"price"`
	PaidPrice          string             `json:"paidPrice"`
	Currency           string             `json:"currency"` // "TRY"
	Installment        int                `json:"installment"`
	BasketID           string             `json:"basketId"`
	PaymentGroup       string             `json:"paymentGroup"` // "PRODUCT" | "SUBSCRIPTION"
	Buyer              CheckoutFormBuyer  `json:"buyer"`
	ShippingAddress    CheckoutFormAddress `json:"shippingAddress"`
	BillingAddress     CheckoutFormAddress `json:"billingAddress"`
	BasketItems        []CheckoutFormItem `json:"basketItems"`
	CallbackURL        string             `json:"callbackUrl"`
	EnabledInstallments []int             `json:"enabledInstallments,omitempty"`
	ThreeDSForce       bool               `json:"threeDSForce"`
}

type CheckoutFormInitResp struct {
	Status        string `json:"status"`        // "success" | "failure"
	Locale        string `json:"locale"`
	SystemTime    int64  `json:"systemTime"`
	ConversationID string `json:"conversationId"`
	Token         string `json:"token"`
	TokenExpire   int64  `json:"tokenExpireTime"` // ms
	CheckoutFormContent string `json:"checkoutFormContent"`
	ErrorCode     string `json:"errorCode"`
	ErrorMessage  string `json:"errorMessage"`
}

// InitCheckoutForm, Iyzico'dan checkout token alır
// (Production'da HTML form content'i embed edilir; geliştirmede stub)
func (c *IyzicoClient) InitCheckoutForm(ctx context.Context, req CheckoutFormReq) (*CheckoutFormInitResp, error) {
	if !c.IsConfigured() {
		// Dev stub: gerçek token üret (UI tarafında mock ödeme akışı)
		fakeToken := fmt.Sprintf("dev-%s-%d", req.BasketID, time.Now().Unix())
		return &CheckoutFormInitResp{
			Status:    "success",
			Token:     fakeToken,
			TokenExpire: time.Now().Add(30 * time.Minute).UnixMilli(),
		}, nil
	}

	body, _ := json.Marshal(req)
	raw, err := c.doSigned(ctx, "/payment/iyzipos/checkoutform/initialize/auth/ecom", body)
	if err != nil {
		return nil, err
	}
	var resp CheckoutFormInitResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("iyzico error %s: %s", resp.ErrorCode, resp.ErrorMessage)
	}
	return &resp, nil
}

type RetrieveCheckoutFormReq struct {
	Locale         string `json:"locale"`
	ConversationID string `json:"conversationId"`
	Token          string `json:"token"`
}

type CheckoutFormRetrieveResp struct {
	Status         string `json:"status"`
	PaymentStatus  string `json:"paymentStatus"` // "SUCCESS" | "FAILURE"
	PaymentID      string `json:"paymentId"`
	PaidPrice      string `json:"paidPrice"`
	Currency       string `json:"currency"`
	Installment    int    `json:"installment"`
	CardFamily     string `json:"cardFamily"`
	CardAssociation string `json:"cardAssociation"`
	CardType       string `json:"cardType"`
	FraudStatus    int    `json:"fraudStatus"`
	ErrorCode      string `json:"errorCode"`
	ErrorMessage   string `json:"errorMessage"`
}

func (c *IyzicoClient) RetrieveCheckoutForm(ctx context.Context, token, conversationID string) (*CheckoutFormRetrieveResp, error) {
	if !c.IsConfigured() {
		// Dev stub: token "dev-" prefixli ise direkt success
		if len(token) > 4 && token[:4] == "dev-" {
			return &CheckoutFormRetrieveResp{
				Status:        "success",
				PaymentStatus: "SUCCESS",
				PaymentID:     fmt.Sprintf("dev-pay-%d", time.Now().Unix()),
			}, nil
		}
		return nil, errors.New("iyzico not configured")
	}

	req := RetrieveCheckoutFormReq{
		Locale:         "tr",
		ConversationID: conversationID,
		Token:          token,
	}
	body, _ := json.Marshal(req)
	raw, err := c.doSigned(ctx, "/payment/iyzipos/checkoutform/auth/ecom/detail", body)
	if err != nil {
		return nil, err
	}
	var resp CheckoutFormRetrieveResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// doSigned, Iyzico'nun HMAC-SHA256 + Authorization header auth şemasını uygular
// Header: Authorization: IYZWSv2 <base64(apiKey + ":" + signature)>
func (c *IyzicoClient) doSigned(ctx context.Context, path string, body []byte) ([]byte, error) {
	uri := c.baseURL + path
	rnd := randomString()

	// signatureString: rnd + uri + body
	sigStr := rnd + uri + string(body)
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(sigStr))
	signature := mac.Sum(nil)

	auth := "IYZWSv2 " + base64.StdEncoding.EncodeToString([]byte(c.apiKey+":"+string(signature)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("x-iyzi-rnd", rnd)
	req.Header.Set("x-iyzi-client-version", "iyzipay-go-1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func randomString() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
		time.Sleep(time.Microsecond)
	}
	return url.QueryEscape(string(b))
}
