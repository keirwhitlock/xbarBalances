package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const starlingApi string = "https://api.starlingbank.com/api/v2"

var cfg = getConfig()

type Config struct {
	Starling struct {
		Token string `yaml:"bearer_token"`
		ID    string `yaml:"account_id"`
	} `yaml:"starling_config"`
}

type AccBalance struct {
	EffectiveBalance struct {
		MinorUnits int    `json:"minorUnits"`
		Currency   string `json:"currency"`
	} `json:"effectiveBalance"`
	TotalEffectiveBalance struct {
		MinorUnits int `json:"minorUnits"`
	} `json:"totalEffectiveBalance"`
}

type Account struct {
	Name string `json:"name"`
	ID   string `json:"accountUid"`
}

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

func callApi(url, token string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func BankAccounts() ([]Account, error) {
	var url string = fmt.Sprintf("%s/accounts", starlingApi)
	content := callApi(url, cfg.Starling.Token)

	var accounts Accounts

	jsonDecoder := json.NewDecoder(strings.NewReader(content))

	err := jsonDecoder.Decode(&accounts)
	if err != nil {
		panic(err)
	}

	return accounts.Accounts, nil
}

func Balance(accountId string) (float64, float64, string, error) {
	var url string = fmt.Sprintf("%s/accounts/%s/balance", starlingApi, accountId)
	content := callApi(url, cfg.Starling.Token)

	var bal AccBalance

	jsonDecoder := json.NewDecoder(strings.NewReader(content))

	err := jsonDecoder.Decode(&bal)
	if err != nil {
		panic(err)
	}

	accountBalance := (float64(bal.EffectiveBalance.MinorUnits) / 100)
	savingsBalance := (float64(bal.TotalEffectiveBalance.MinorUnits-bal.EffectiveBalance.MinorUnits) / 100)
	var currency string

	switch bal.EffectiveBalance.Currency {
	case "GBP":
		currency = "£"
	case "EUR":
		currency = "€"
	case "USD":
		currency = "$"
	default:
		currency = "?"
	}

	return accountBalance, savingsBalance, currency, nil
}

func getConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	f, err := os.Open(fmt.Sprintf("%s/development/go/starlingBalance/config.yml", homeDir))
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var cfg Config

	yamlDecoder := yaml.NewDecoder(f)
	err = yamlDecoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func main() {

	accounts, err := BankAccounts()
	if err != nil {
		panic(err)
	}

	for _, account := range accounts {
		bal, sav, cur, err := Balance(account.ID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Starling - %[1]s bal: %[2]s%[3]v (sav: %[2]s%[4]v)\n", account.Name, cur, bal, sav)
	}
}
