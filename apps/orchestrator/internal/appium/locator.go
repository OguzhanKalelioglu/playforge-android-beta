package appium

import (
	"context"
	"fmt"
	"net/http"
)

// By, locator stratejisi
type By string

const (
	ByID          By = "id"           // resource-id
	ByAccessibilityID By = "accessibility id"  // content-desc
	ByClassName   By = "class name"
	ByXPath       By = "xpath"
	ByName        By = "name"         // text
	ByAndroidUIAutomator By = "-android uiautomator"
	ByCSS         By = "css selector"
)

// Element, WebDriver element handle
type Element struct {
	session *Session
	eid     string
}

// FindElement, tek bir element bulur
func (s *Session) FindElement(ctx context.Context, by By, value string) (*Element, error) {
	if err := s.checkActive(); err != nil {
		return nil, err
	}

	req := map[string]interface{}{
		"using": string(by),
		"value": value,
	}

	var resp struct {
		Value struct {
			ELEMENT string `json:"ELEMENT"`
			ElementID string `json:"element-6066-521e-4f87-b2c0-2693eb248f5e"` // W3C
		} `json:"value"`
	}
	if err := s.client.do(ctx, http.MethodPost, s.path("/element"), req, &resp); err != nil {
		return nil, err
	}

	eid := resp.Value.ElementID
	if eid == "" {
		eid = resp.Value.ELEMENT
	}
	if eid == "" {
		return nil, fmt.Errorf("element not found (by=%s, value=%s)", by, value)
	}
	return &Element{session: s, eid: eid}, nil
}

// FindElements, birden fazla element bulur
func (s *Session) FindElements(ctx context.Context, by By, value string) ([]*Element, error) {
	if err := s.checkActive(); err != nil {
		return nil, err
	}

	req := map[string]interface{}{
		"using": string(by),
		"value": value,
	}

	var resp struct {
		Value []struct {
			ELEMENT string `json:"ELEMENT"`
			ElementID string `json:"element-6066-521e-4f87-b2c0-2693eb248f5e"`
		} `json:"value"`
	}
	if err := s.client.do(ctx, http.MethodPost, s.path("/elements"), req, &resp); err != nil {
		return nil, err
	}

	out := make([]*Element, 0, len(resp.Value))
	for _, v := range resp.Value {
		eid := v.ElementID
		if eid == "" {
			eid = v.ELEMENT
		}
		if eid != "" {
			out = append(out, &Element{session: s, eid: eid})
		}
	}
	return out, nil
}

func (e *Element) path(suffix string) string {
	return e.session.path("/element/" + e.eid + suffix)
}

// Text, element'in text içeriğini döndürür
func (e *Element) Text(ctx context.Context) (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := e.session.client.do(ctx, http.MethodGet, e.path("/text"), nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// Attribute, element'in attribute değerini döndürür
func (e *Element) Attribute(ctx context.Context, name string) (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := e.session.client.do(ctx, http.MethodGet, e.path("/attribute/"+name), nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// IsDisplayed, element görünür mü?
func (e *Element) IsDisplayed(ctx context.Context) (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := e.session.client.do(ctx, http.MethodGet, e.path("/displayed"), nil, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// IsEnabled, element aktif mi?
func (e *Element) IsEnabled(ctx context.Context) (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := e.session.client.do(ctx, http.MethodGet, e.path("/enabled"), nil, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}
