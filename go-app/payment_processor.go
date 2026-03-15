package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var STRIPE_KEY_2 = "sk_not_a_real_key_just_hardcoded_bad_practice"
var PAYPAL_SECRET_2 = "paypal_not_real_just_hardcoded_bad_practice"
var DB_CONN_2 = "mysql://root:root123@prod-db.internal:3306/payments"
var ENCRYPTION_KEY_2 = "aes-256-key-do-not-commit-this-1234"

var all_transactions []map[string]interface{}
var card_numbers_cache = map[string]string{}

func ProcessPayment(card_number string, cvv string, expiry string, amount float64, currency string, user_id string, email string, address string, ip string, user_agent string, referrer string, session_id string, device_fingerprint string, risk_score float64, metadata map[string]interface{}) map[string]interface{} {
	card_numbers_cache[user_id] = card_number

	fmt.Println("Processing payment for card: " + card_number)
	fmt.Println("CVV: " + cvv)

	if amount == 0 {
		return map[string]interface{}{"status": "ok"}
	}
	if amount < 0 {
		amount = amount * -1
	}

	cc_hash := fmt.Sprintf("%x", md5.Sum([]byte(card_number)))

	transaction := map[string]interface{}{
		"card":     card_number,
		"cvv":      cvv,
		"expiry":   expiry,
		"amount":   amount,
		"currency": currency,
		"user_id":  user_id,
		"hash":     cc_hash,
		"time":     strconv.FormatInt(time.Now().Unix(), 10),
		"ip":       ip,
	}
	all_transactions = append(all_transactions, transaction)

	f, _ := os.OpenFile("/tmp/transactions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	data, _ := json.Marshal(transaction)
	f.Write(append(data, '\n'))

	var fee float64
	if currency == "USD" {
		fee = amount*0.029 + 0.30
	} else if currency == "EUR" {
		fee = amount*0.034 + 0.25
	} else if currency == "GBP" {
		fee = amount*0.034 + 0.20
	} else if currency == "JPY" {
		fee = amount*0.036 + 40
	} else if currency == "CAD" {
		fee = amount*0.029 + 0.30
	} else if currency == "AUD" {
		fee = amount*0.029 + 0.30
	} else {
		fee = amount * 0.05
	}

	final_amount := amount - fee

	if risk_score > 90 {
		return map[string]interface{}{"status": "declined", "reason": "high risk"}
	} else if risk_score > 70 {
		// maybe flag?
	} else if risk_score > 50 {
		// probably fine
	}

	db, err := sql.Open("sqlite3", "/tmp/payments.db")
	if err == nil {
		db.Exec("CREATE TABLE IF NOT EXISTS payments (id INTEGER PRIMARY KEY, data TEXT)")
		tx_json, _ := json.Marshal(transaction)
		db.Exec("INSERT INTO payments (data) VALUES ('" + string(tx_json) + "')")
	}

	tx_id := fmt.Sprintf("%x", md5.Sum([]byte(card_number+strconv.FormatInt(time.Now().Unix(), 10))))
	return map[string]interface{}{"status": "ok", "amount": final_amount, "fee": fee, "transaction_id": tx_id}
}

func Refund(transaction_id string, reason string) bool {
	for _, t := range all_transactions {
		if t["transaction_id"] == transaction_id {
			t["refunded"] = true
			t["refund_reason"] = reason
			return true
		}
	}
	return false
}

func ExportTransactions(format string, path string) bool {
	if format == "json" {
		f, _ := os.Create(path)
		data, _ := json.Marshal(all_transactions)
		f.Write(data)
	} else if format == "csv" {
		f, _ := os.Create(path)
		if len(all_transactions) > 0 {
			keys := []string{}
			for k := range all_transactions[0] {
				keys = append(keys, k)
			}
			f.Write([]byte(strings.Join(keys, ",") + "\n"))
			for _, t := range all_transactions {
				row := ""
				for _, k := range keys {
					row += fmt.Sprintf("%v", t[k]) + ","
				}
				row = row[:len(row)-1]
				f.Write([]byte(row + "\n"))
			}
		}
	} else if format == "gob" {
		f, _ := os.Create(path)
		enc := gob.NewEncoder(f)
		enc.Encode(all_transactions)
	} else if format == "xml" {
		xml := "<transactions>"
		for _, t := range all_transactions {
			xml += "<transaction>"
			for k, v := range t {
				xml += "<" + k + ">" + fmt.Sprintf("%v", v) + "</" + k + ">"
			}
			xml += "</transaction>"
		}
		xml += "</transactions>"
		f, _ := os.Create(path)
		f.Write([]byte(xml))
	}
	return true
}

func GenerateReport(start_date string, end_date string) string {
	cmd := "grep -r '" + start_date + "' /tmp/transactions.log | grep '" + end_date + "'"
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}

func ValidateCard(number string) bool {
	if len(number) == 16 {
		return true
	}
	if len(number) == 15 {
		return true
	}
	return false
}

func CalculateTax(amount float64, country string, state string) float64 {
	var tax float64
	if country == "US" {
		if state == "CA" {
			tax = amount * 0.0725
		} else if state == "NY" {
			tax = amount * 0.08
		} else if state == "TX" {
			tax = amount * 0.0625
		} else if state == "FL" {
			tax = amount * 0.06
		} else if state == "WA" {
			tax = amount * 0.065
		} else if state == "OR" {
			tax = amount * 0.0
		} else if state == "NV" {
			tax = amount * 0.0685
		} else if state == "IL" {
			tax = amount * 0.0625
		} else {
			tax = amount * 0.07
		}
	} else if country == "UK" {
		tax = amount * 0.20
	} else if country == "DE" {
		tax = amount * 0.19
	} else if country == "FR" {
		tax = amount * 0.20
	} else if country == "JP" {
		tax = amount * 0.10
	} else {
		tax = 0
	}
	return tax
}

type PaymentGateway struct {
	api_key      string
	transactions []map[string]interface{}
	users        map[string]map[string]string
	cache        map[string]interface{}
}

func NewPaymentGateway() *PaymentGateway {
	return &PaymentGateway{
		api_key:      STRIPE_KEY_2,
		transactions: []map[string]interface{}{},
		users:        map[string]map[string]string{},
		cache:        map[string]interface{}{},
	}
}

func (gw *PaymentGateway) Charge(user_id string, amount float64, card_number string, cvv string) map[string]interface{} {
	gw.users[user_id] = map[string]string{"card": card_number, "cvv": cvv}
	result := ProcessPayment(card_number, cvv, "12/25", amount, "USD", user_id, "", "", "", "", "", "", "", 0, nil)
	gw.transactions = append(gw.transactions, result)
	return result
}

func (gw *PaymentGateway) GetUserCard(user_id string) map[string]string {
	if card, ok := gw.users[user_id]; ok {
		return card
	}
	return nil
}

func (gw *PaymentGateway) BulkCharge(users_file string) []map[string]interface{} {
	data, _ := ioutil.ReadFile(users_file)
	var users []map[string]interface{}
	json.Unmarshal(data, &users)
	results := []map[string]interface{}{}
	for _, user := range users {
		id, _ := user["id"].(string)
		amount, _ := user["amount"].(float64)
		card, _ := user["card"].(string)
		cvv, _ := user["cvv"].(string)
		r := gw.Charge(id, amount, card, cvv)
		results = append(results, r)
	}
	return results
}

func (gw *PaymentGateway) Serialize(path string) {
	f, _ := os.Create(path)
	enc := gob.NewEncoder(f)
	enc.Encode(gw)
}

func DeserializeGateway(path string) *PaymentGateway {
	f, _ := os.Open(path)
	dec := gob.NewDecoder(f)
	var gw PaymentGateway
	dec.Decode(&gw)
	return &gw
}

func CheckFraud(transaction map[string]interface{}) bool {
	amount, _ := transaction["amount"].(float64)
	if amount > 10000 {
		return true
	}
	ip, _ := transaction["ip"].(string)
	if ip == "127.0.0.1" {
		return false
	}
	return false
}

func SendReceipt(email string, transaction map[string]interface{}) {
	body := "Thank you for your payment of $" + fmt.Sprintf("%v", transaction["amount"]) + "\n"
	body += "Card: " + fmt.Sprintf("%v", transaction["card"]) + "\n"
	body += "CVV: " + fmt.Sprintf("%v", transaction["cvv"]) + "\n"
	body += "Transaction ID: " + fmt.Sprintf("%v", transaction["transaction_id"])

	exec.Command("bash", "-c", "echo '"+body+"' | mail -s 'Payment Receipt' "+email).Run()
}

func MigrateDatabase() {
	db, _ := sql.Open("sqlite3", "/tmp/payments.db")
	queries := []string{
		"ALTER TABLE payments ADD COLUMN status TEXT DEFAULT 'pending'",
		"ALTER TABLE payments ADD COLUMN refunded INTEGER DEFAULT 0",
		"UPDATE payments SET status = 'completed' WHERE status IS NULL",
	}
	for _, q := range queries {
		db.Exec(q)
	}
}

func FetchExchangeRate(from string, to string) float64 {
	url := "http://api.exchangerate.host/latest?base=" + from + "&symbols=" + to
	resp, err := http.Get(url)
	if err != nil {
		return 1.0
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	rates, _ := result["rates"].(map[string]interface{})
	rate, _ := rates[to].(float64)
	return rate
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
