package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cart = Cart{
	Reference: "2d832fe0-6c96-4515-9be7-4c00983539c1",
	LineItems: []LineItem{
		{
			Name:  "Peanut Butter",
			Price: 39.0,
			Sku:   "PEANUT-BUTTER",
		},
		{
			Name:  "Fruity",
			Price: 34.99,
			Sku:   "FRUITY",
		},
		{
			Name:  "Chocolate",
			Price: 32,
			Sku:   "CHOCOLATE",
		},
	},
}

var cart_multiple_eligible = Cart{
	Reference: "2d832fe0-6c96-4515-9be7-4c00983539c1",
	LineItems: []LineItem{
		{
			Name:  "Peanut Butter",
			Price: 39.0,
			Sku:   "PEANUT-BUTTER",
		},
		{
			Name:  "Cocoa",
			Price: 35.,
			Sku:   "COCOA",
		},
		{
			Name:  "Chocolate",
			Price: 32.,
			Sku:   "CHOCOLATE",
		},
		{
			Name:  "Banana Cake",
			Price: 36.,
			Sku:   "BANANA-CAKE",
		},
	},
}

var cart_not_eligible = Cart{
	Reference: "2d832fe0-6c96-4515-9be7-4c00983539c1",
	LineItems: []LineItem{
		{
			Name:  "Banana Cake",
			Price: 36.,
			Sku:   "BANANA-CAKE",
		},
		{
			Name:  "Chocolate",
			Price: 32,
			Sku:   "CHOCOLATE",
		},
	},
}

func TestCalculateCartDiscount(t *testing.T) {
	cashier := NewCashier(config)

	tests := []struct {
		name     string
		cart     Cart
		expected float64
	}{
		{
			name:     "default cart gives expected result",
			cart:     cart,
			expected: 89.99,
		},
		{
			name:     "multiple elegible products only one discount on cheapest",
			cart:     cart_multiple_eligible,
			expected: 126.,
		},
		{
			name:     "no elegible products no discount",
			cart:     cart_not_eligible,
			expected: 68.,
		},
	}

	for i, testCase := range tests {
		actual, err := cashier.CalculateTotalWithDiscount(testCase.cart)
		assert.NoError(t, err)

		if !assert.Equal(t, testCase.expected, actual.Total) {
			t.Errorf("failed at idx %d, test name: %s, got: %v, expected: %v", i, testCase.name, actual.Total, testCase.expected)
		}
	}
}

func TestGetEligibleForDiscount(t *testing.T) {
	cashier := NewCashier(config)

	tests := []struct {
		name      string
		cartItems []LineItem
		expected  []LineItem
	}{
		{
			name:      "default cart with chocolate",
			cartItems: cart.LineItems,
			expected: []LineItem{
				{
					Name:  "Chocolate",
					Sku:   "CHOCOLATE",
					Price: 32,
				},
			},
		},
		{
			name: "no eligible items",
			cartItems: []LineItem{{
				Sku: "another-sku",
			}},
			expected: []LineItem{},
		},
		{
			name: "multiple items",
			cartItems: []LineItem{
				{
					Sku: "CHOCOLATE",
				},
				{
					Sku: "PEANUT-BUTTER",
				},
				{
					Sku: "FRUITY",
				},
				{
					Sku: "FRUITY",
				},
			},
			expected: []LineItem{
				{
					Sku: "CHOCOLATE",
				},
			},
		},
		{
			name: "entire cart",
			cartItems: []LineItem{
				{
					Sku: "BANANA-CAKE",
				},
				{
					Sku: "COCOA",
				},
				{
					Sku: "CHOCOLATE",
				},
			},
			expected: []LineItem{
				{
					Sku: "BANANA-CAKE",
				},
				{
					Sku: "COCOA",
				},
				{
					Sku: "CHOCOLATE",
				},
			},
		},
	}

	for i, testCase := range tests {
		actual := cashier.getEligibleForDiscount(testCase.cartItems)
		if !assert.True(t, slices.Equal(actual, testCase.expected)) {
			t.Errorf("failed at idx %d, test name: %s, got: %v, expected: %v", i, testCase.name, actual, testCase.expected)
		}
	}
}

func TestHasPrerequisiteSkus(t *testing.T) {
	cashier := NewCashier(config)

	tests := []struct {
		cartItems []LineItem
		expected  bool
	}{
		{
			cartItems: cart.LineItems,
			expected:  true,
		},
		{
			cartItems: []LineItem{},
			expected:  false,
		},
		{
			cartItems: []LineItem{{
				Sku: "another-sku",
			}},
			expected: false,
		},
		{
			cartItems: []LineItem{{
				Sku: "PEANUT-BUTTER",
			}},
			expected: true,
		},
	}

	for _, testCase := range tests {
		actual := cashier.hasPrerequisiteSku(testCase.cartItems)
		assert.Equal(t, testCase.expected, actual)
	}

}

func TestGetCheapestProduct(t *testing.T) {
	cashier := NewCashier(config)

	tests := []struct {
		name      string
		cartItems []LineItem
		expected  *LineItem
	}{
		{
			name:      "default items",
			cartItems: cart.LineItems,
			expected: &LineItem{
				Name:  "Chocolate",
				Price: 32,
				Sku:   "CHOCOLATE",
			},
		},
		{
			cartItems: []LineItem{},
			expected:  nil,
		},
	}

	for i, testCase := range tests {
		actual, err := cashier.getCheapestProduct(testCase.cartItems)
		if testCase.expected != nil {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		if !assert.Equal(t, testCase.expected, actual) {
			t.Errorf("failed at idx %d, test name: %s, got: %v, expected: %v", i, testCase.name, actual, testCase.expected)
		}
	}

}

func TestGetDiscount(t *testing.T) {
	cashier := NewCashier(config)

	tests := []struct {
		expected float64
		amount   float64
	}{
		{
			expected: 16,
			amount:   32,
		},
	}

	for _, testCase := range tests {
		actual, err := cashier.calculateDiscount(testCase.amount)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expected, actual)
	}
}

// http test
func TestCalculateTotalWithDiscountEndpoint(t *testing.T) {
	cart := cart

	body, _ := json.Marshal(CartRequest{Cart: cart})

	request := httptest.NewRequest(http.MethodPost, "/cart/total", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(shoppingCartHandler)
	handler.ServeHTTP(rr, request)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var cartResponse Cart
	if err := json.Unmarshal(rr.Body.Bytes(), &cartResponse); err != nil {
		t.Fatalf("could not parse response body: %v", err)
	}

	if cartResponse.Total <= 0 {
		t.Errorf("expected total > 0, got: %v", cartResponse)
	}

	if cartResponse.Total != 89.99 {
		t.Errorf("incorrect total, got: %v", cartResponse)
	}
}
