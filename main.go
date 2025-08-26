package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var config = Configuration{
	PrerequisiteSkus: []string{"PEANUT-BUTTER", "COCOA", "FRUITY"},
	EligibleSkus:     []string{"BANANA-CAKE", "COCOA", "CHOCOLATE"},
	DiscountUnit:     "percentage",
	DiscountValue:    50.0,
}

type Configuration struct {
	PrerequisiteSkus []string `json:"prerequisite_skus"`
	EligibleSkus     []string `json:"eligible_skus"`
	DiscountUnit     string   `json:"discount_unit"`
	DiscountValue    float64  `json:"discount_value"`
}

type Cart struct {
	Reference string     `json:"reference"`
	LineItems []LineItem `json:"lineItems"`
	Total     float64    `json:"total"`
}

type LineItem struct {
	Name            string  `json:"name"`
	Price           float64 `json:"price"`
	Sku             string  `json:"sku"`
	DiscountedPrice float64 `json:"discountedPrice"`
}

type Cashier struct {
	config Configuration
}

type CartRequest struct {
	Cart Cart `json:"cart"`
}

func NewCashier(config Configuration) *Cashier {
	return &Cashier{
		config: config,
	}
}

func (c *Cashier) CalculateTotalWithDiscount(cart Cart) (Cart, error) {
	discountedItems := make([]LineItem, len(cart.LineItems))
	copy(discountedItems, cart.LineItems)

	var total float64

	hasPrerequisite := c.hasPrerequisiteSku(discountedItems)
	var eligibleItems []LineItem
	if hasPrerequisite {
		eligibleItems = c.getEligibleForDiscount(discountedItems)
	}

	var discountedSKU string
	var discountValue float64
	if len(eligibleItems) > 0 {
		cheapest, err := c.getCheapestProduct(eligibleItems)
		if err != nil {
			return Cart{}, err
		}
		discountAmount, err := c.calculateDiscount(cheapest.Price)
		if err != nil {
			return Cart{}, err
		}
		discountedSKU = cheapest.Sku
		discountValue = discountAmount
	}

	for i, item := range discountedItems {
		itemDiscount := 0.0
		if item.Sku == discountedSKU {
			itemDiscount = discountValue
		}

		discountedItems[i].DiscountedPrice = c.parseFloat(item.Price - itemDiscount)
		total += discountedItems[i].DiscountedPrice
	}

	return Cart{
		Reference: cart.Reference,
		LineItems: discountedItems,
		Total:     c.parseFloat(total),
	}, nil
}

func (c *Cashier) parseFloat(amount float64) float64 {
	formatted := fmt.Sprintf("%.2f", amount)
	parsed, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		return -1.
	}

	return parsed
}

func (c *Cashier) getEligibleForDiscount(items []LineItem) []LineItem {
	eligibleSkus := make(map[string]bool)

	for _, sku := range c.config.EligibleSkus {
		eligibleSkus[sku] = true
	}

	var eligibleForDiscount []LineItem
	for _, item := range items {
		if eligibleSkus[item.Sku] {
			eligibleForDiscount = append(eligibleForDiscount, item)
		}
	}

	return eligibleForDiscount
}

func (c *Cashier) hasPrerequisiteSku(items []LineItem) bool {
	prerequisiteSkus := make(map[string]bool)

	for _, sku := range c.config.PrerequisiteSkus {
		prerequisiteSkus[sku] = true
	}

	for _, item := range items {
		if prerequisiteSkus[item.Sku] {
			return true
		}
	}

	return false
}

func (c *Cashier) getCheapestProduct(items []LineItem) (*LineItem, error) {
	if len(items) < 1 {
		return nil, errors.New("empty list of products")
	}

	cheapest := items[0]

	for _, item := range items {
		if cheapest.Price > item.Price {
			cheapest = item
		}
	}

	return &cheapest, nil
}

func (c *Cashier) calculateDiscount(amount float64) (float64, error) {
	switch c.config.DiscountUnit {
	case "percentage":
		return amount * c.config.DiscountValue / 100, nil
	default:
		return -1, errors.New("unsupported discount method")
	}
}

func main() {
	http.HandleFunc("POST /cart/total", shoppingCartHandler)

	fmt.Println("server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func shoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	var cartRequest CartRequest
	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if len(cartRequest.Cart.LineItems) == 0 {
		http.Error(w, "empty cart", http.StatusBadRequest)
		return
	}

	cashier := NewCashier(config)

	newCart, err := cashier.CalculateTotalWithDiscount(cartRequest.Cart)
	if err != nil {
		http.Error(w, "invalid cart data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCart)
}
