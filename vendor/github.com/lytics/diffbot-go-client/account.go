package diffbot

import (
	"encoding/json"
	"net/http"
)

// See http://diffbot.com/dev/docs/account/
type Account struct {
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Plan        string     `json:"plan"`
	PlanCalls   int        `json:"planCalls"`
	Status      string     `json:"status"`
	ChildTokens []string   `json:"childTokens"`
	ApiCalls    []*apiCall `json:"apiCalls"`
	Invoices    []*invoice `json:"invoices"`
}

type apiCall struct {
	Date       string `json:"date"`
	Calls      int    `json:"calls"`
	ProxyCalls int    `json:"proxyCalls"`
	GiCalls    int    `json:"giCalls"`
}

type invoice struct {
	Date          string  `json:"date"`
	PeriodStart   string  `json:"periodStart"`
	PeriodEnd     string  `json:"periodEnd"`
	TotalCalls    int     `json:"totalCalls"`
	TotalAmount   float64 `json:"totalAmount"`
	OverageAmount float64 `json:"overageAmount"`
	Status        string  `json:"status"`
}

func GetAccount(client *http.Client, token string) (*Account, error) {
	body, err := Diffbot(client, "account", token, "", &Options{})
	if err != nil {
		return nil, err
	}

	account := &Account{}
	if err := json.Unmarshal(body, account); err != nil {
		return nil, err
	}
	return account, nil
}
