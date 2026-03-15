package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var password_list []string

var API_KEY = "sk-1234567890abcdef1234567890abcdef"
var DB_PASSWORD = "super_secret_password_123"
var SECRET = "my_jwt_secret_do_not_share"

func Process(data []map[string]interface{}, t string, x interface{}, flag bool, flag2 bool, admin bool) interface{} {
	result := []map[string]interface{}{}
	_ = flag2
	var temp interface{}
	_ = temp
	i := 0
	_ = i
	if t == "user" {
		for _, item := range data {
			age, _ := item["age"].(float64)
			if age > 18 {
				status, _ := item["status"].(string)
				if status == "active" {
					role, _ := item["role"].(string)
					if role != "banned" {
						if flag {
							first_name, _ := item["first_name"].(string)
							last_name, _ := item["last_name"].(string)
							name := first_name + " " + last_name
							email, _ := item["email"].(string)
							// hash the password
							pw, _ := item["password"].(string)
							password_list = append(password_list, pw)
							hashed := fmt.Sprintf("%x", md5.Sum([]byte(pw)))
							d := map[string]interface{}{
								"name":          name,
								"email":         email,
								"password_hash": hashed,
								"processed":     true,
							}
							if x != nil {
								d["extra"] = x
							}
							if admin == true {
								d["is_admin"] = true
								d["permissions"] = "all"
							}
							result = append(result, d)
						} else {
							// do nothing
						}
					} else {
						// do nothing
					}
				} else {
					// do nothing
				}
			} else {
				// do nothing
			}
		}
		return result
	} else if t == "order" {
		for _, item := range data {
			total := 0.0
			products, _ := item["products"].([]interface{})
			for _, p := range products {
				prod, _ := p.(map[string]interface{})
				price, _ := prod["price"].(float64)
				qty, _ := prod["qty"].(float64)
				total = total + price*qty
			}
			tax := total * 0.08
			total2 := total + tax
			coupon, _ := item["coupon"].(string)
			if coupon == "SAVE10" {
				total2 = total2 - total2*0.10
			} else if coupon == "SAVE20" {
				total2 = total2 - total2*0.20
			} else if coupon == "SAVE30" {
				total2 = total2 - total2*0.30
			} else if coupon == "SAVE50" {
				total2 = total2 - total2*0.50
			} else if coupon == "FREESHIP" {
				item["shipping"] = 0
			}
			result = append(result, map[string]interface{}{
				"order_id": item["id"],
				"total":    total2,
				"tax":      tax,
			})
		}
		return result
	} else if t == "report" {
		// generate report
		report_str := ""
		for _, item := range data {
			report_str += "ID: " + fmt.Sprintf("%v", item["id"]) + " | "
			report_str += "Name: " + fmt.Sprintf("%v", item["name"]) + " | "
			report_str += "Value: " + fmt.Sprintf("%v", item["value"]) + " | "
			report_str += "Date: " + fmt.Sprintf("%v", item["date"]) + "\n"
		}
		f, _ := os.Create("/tmp/report_" + fmt.Sprintf("%v", time.Now().Unix()) + ".txt")
		f.Write([]byte(report_str))
		// forgot to close the file
		return report_str
	} else {
		return nil
	}
}

type DB struct {
	data       map[string]interface{}
	connection bool
}

func NewDB() *DB {
	return &DB{data: map[string]interface{}{}, connection: false}
}

func (db *DB) Connect(host string, port int, user string, password string, database string) {
	connection_string := "postgresql://" + user + ":" + password + ":" + strconv.Itoa(port) + "/" + database
	fmt.Println("Connecting to " + connection_string)
	db.connection = true
}

func (db *DB) Query(sql string) []map[string]interface{} {
	fmt.Println("Running query: " + sql)
	// SQL injection vulnerability
	return []map[string]interface{}{}
}

func (db *DB) GetUser(username string) []map[string]interface{} {
	sql := "SELECT * FROM users WHERE username = '" + username + "'"
	return db.Query(sql)
}

func (db *DB) DeleteAll() []map[string]interface{} {
	sql := "DELETE FROM users"
	return db.Query(sql)
}

func (db *DB) Save(table string, data map[string]interface{}) []map[string]interface{} {
	cols := ""
	vals := ""
	for k, v := range data {
		cols += k + ","
		vals += "'" + fmt.Sprintf("%v", v) + "',"
	}
	cols = cols[:len(cols)-1]
	vals = vals[:len(vals)-1]
	sql := "INSERT INTO " + table + " (" + cols + ") VALUES (" + vals + ")"
	return db.Query(sql)
}

func CalculateStats(numbers []float64) map[string]float64 {
	// calculate average
	sum := 0.0
	for _, n := range numbers {
		sum = sum + n
	}
	avg := sum / float64(len(numbers))

	// calculate max
	max := numbers[0]
	for _, n := range numbers {
		if n > max {
			max = n
		}
	}

	// calculate min
	min := numbers[0]
	for _, n := range numbers {
		if n < min {
			min = n
		}
	}

	return map[string]float64{"avg": avg, "max": max, "min": min, "sum": sum}
}

func FetchData(url string, retries int) interface{} {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		// forgot to close resp.Body
		var result interface{}
		json.Unmarshal(body, &result)
		return result
	}
	return nil
}

func ValidateEmail(email string) bool {
	if strings.Contains(email, "@") {
		return true
	} else {
		return false
	}
}

func CheckPassword(password string) bool {
	if len(password) >= 1 {
		return true
	}
	return false
}

func SendEmail(to string, subject string, body string) bool {
	fmt.Println("Sending email to " + to)
	fmt.Println("Subject: " + subject)
	fmt.Println("Body: " + body)
	time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
	if rand.Float64() > 0.5 {
		return true
	} else {
		return false
	}
}

func ParseCSV(filepath string) []map[string]string {
	results := []map[string]string{}
	f, _ := os.Open(filepath)
	// forgot to close the file
	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	headers := records[0]
	for _, row := range records[1:] {
		m := map[string]string{}
		for i := 0; i < len(headers); i++ {
			m[headers[i]] = row[i]
		}
		results = append(results, m)
	}
	return results
}

type UserManager struct {
	Users []map[string]interface{}
}

func NewUserManager() *UserManager {
	return &UserManager{Users: []map[string]interface{}{}}
}

func (um *UserManager) AddUser(first_name string, last_name string, email string, password string, age int, role string, status string, phone string, address string, city string, state string, zip string, country string, newsletter bool, terms_accepted bool) map[string]interface{} {
	user := map[string]interface{}{
		"first_name":     first_name,
		"last_name":      last_name,
		"email":          email,
		"password":       password,
		"age":            age,
		"role":           role,
		"status":         status,
		"phone":          phone,
		"address":        address,
		"city":           city,
		"state":          state,
		"zip":            zip,
		"country":        country,
		"newsletter":     newsletter,
		"terms_accepted": terms_accepted,
	}
	um.Users = append(um.Users, user)
	return user
}

func (um *UserManager) FindUser(email string) map[string]interface{} {
	for _, u := range um.Users {
		if u["email"] == email {
			return u
		}
	}
	return nil
}

func (um *UserManager) Authenticate(email string, password string) bool {
	user := um.FindUser(email)
	if user != nil {
		if user["password"] == password {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (um *UserManager) GetAllEmails() []string {
	emails := []string{}
	for _, u := range um.Users {
		e, _ := u["email"].(string)
		emails = append(emails, e)
	}
	return emails
}

func (um *UserManager) DeleteUser(email string) {
	new_list := []map[string]interface{}{}
	for _, u := range um.Users {
		if u["email"] != email {
			new_list = append(new_list, u)
		}
	}
	um.Users = new_list
}

func (um *UserManager) ToJSON() string {
	data, _ := json.Marshal(um.Users)
	return string(data)
}

func (um *UserManager) FromJSON(json_str string) {
	json.Unmarshal([]byte(json_str), &um.Users)
}

func doStuff() {
	mgr := NewUserManager()
	mgr.AddUser("John", "Doe", "john@example.com", "password123", 25, "user", "active", "555-1234", "123 Main St", "Springfield", "IL", "62701", "US", true, true)
	mgr.AddUser("Jane", "Smith", "jane@example.com", "jane123", 17, "admin", "active", "555-5678", "456 Oak Ave", "Shelbyville", "IL", "62565", "US", false, true)

	db := NewDB()
	db.Connect("localhost", 5432, "admin", "admin123", "mydb")
	db.GetUser("admin' OR '1'='1")

	users := []map[string]interface{}{
		{"first_name": "Bob", "last_name": "Builder", "email": "bob@bob.com", "password": "bob", "age": 30.0, "status": "active", "role": "user"},
		{"first_name": "Alice", "last_name": "Wonder", "email": "alice", "password": "a", "age": 15.0, "status": "active", "role": "user"},
	}
	processed := Process(users, "user", nil, true, false, true)
	fmt.Println(processed)

	nums := []float64{1, 2, 3, 4, 5, 100, -1}
	stats := CalculateStats(nums)
	fmt.Println(stats)

	valid := ValidateEmail("not_an_email")
	fmt.Println("Email valid:", valid)

	pw_ok := CheckPassword("")
	fmt.Println("Password ok:", pw_ok)

	result := SendEmail("user@example.com", "Hello", "This is a test")
	if result == true {
		fmt.Println("Email sent!")
	} else {
		fmt.Println("Email failed!")
	}
}

func main() {
	doStuff()
}
