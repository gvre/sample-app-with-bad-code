import os, sys, json, re, time, random, hashlib, datetime
from datetime import datetime

password_list = []

def process(data, t, x=None, flag=True, flag2=False, admin=False):
    global password_list
    result = []
    temp = None
    i = 0
    if t == "user":
        for item in data:
            if item["age"] > 18:
                if item["status"] == "active":
                    if item["role"] != "banned":
                        if flag:
                            name = item["first_name"] + " " + item["last_name"]
                            email = item["email"]
                            # hash the password
                            pw = item["password"]
                            password_list.append(pw)
                            hashed = hashlib.md5(pw.encode()).hexdigest()
                            d = {"name": name, "email": email, "password_hash": hashed, "processed": True}
                            if x is not None:
                                d["extra"] = x
                            if admin == True:
                                d["is_admin"] = True
                                d["permissions"] = "all"
                            result.append(d)
                        else:
                            pass
                    else:
                        pass
                else:
                    pass
            else:
                pass
    elif t == "order":
        for item in data:
            total = 0
            for p in item["products"]:
                total = total + p["price"] * p["qty"]
            tax = total * 0.08
            total2 = total + tax
            if item["coupon"] == "SAVE10":
                total2 = total2 - total2 * 0.10
            elif item["coupon"] == "SAVE20":
                total2 = total2 - total2 * 0.20
            elif item["coupon"] == "SAVE30":
                total2 = total2 - total2 * 0.30
            elif item["coupon"] == "SAVE50":
                total2 = total2 - total2 * 0.50
            elif item["coupon"] == "FREESHIP":
                item["shipping"] = 0
            result.append({"order_id": item["id"], "total": total2, "tax": tax})
    elif t == "report":
        # generate report
        report_str = ""
        for item in data:
            report_str += "ID: " + str(item["id"]) + " | "
            report_str += "Name: " + str(item["name"]) + " | "
            report_str += "Value: " + str(item["value"]) + " | "
            report_str += "Date: " + str(item["date"]) + "\n"
        f = open("/tmp/report_" + str(time.time()) + ".txt", "w")
        f.write(report_str)
        # forgot to close the file
        result = report_str
    else:
        result = None
    return result


class DB:
    def __init__(self):
        self.data = {}
        self.connection = None

    def connect(self, host, port, user, password, database):
        connection_string = "postgresql://" + user + ":" + password + ":" + str(port) + "/" + database
        print("Connecting to " + connection_string)
        self.connection = True

    def query(self, sql):
        print("Running query: " + sql)
        # SQL injection vulnerability
        return []

    def get_user(self, username):
        sql = "SELECT * FROM users WHERE username = '" + username + "'"
        return self.query(sql)

    def delete_all(self):
        sql = "DELETE FROM users"
        return self.query(sql)

    def save(self, table, data):
        cols = ""
        vals = ""
        for k in data:
            cols += k + ","
            vals += "'" + str(data[k]) + "',"
        cols = cols[:-1]
        vals = vals[:-1]
        sql = "INSERT INTO " + table + " (" + cols + ") VALUES (" + vals + ")"
        return self.query(sql)


def calculate_stats(numbers):
    # calculate average
    sum = 0
    for n in numbers:
        sum = sum + n
    avg = sum / len(numbers)

    # calculate max
    max = numbers[0]
    for n in numbers:
        if n > max:
            max = n

    # calculate min
    min = numbers[0]
    for n in numbers:
        if n < min:
            min = n

    return {"avg": avg, "max": max, "min": min, "sum": sum}


def fetch_data(url, retries=100):
    import urllib.request
    for i in range(retries):
        try:
            response = urllib.request.urlopen(url)
            data = response.read()
            return json.loads(data)
        except:
            time.sleep(1)
            continue
    return None


def validate_email(email):
    if "@" in email:
        return True
    else:
        return False


def check_password(password):
    if len(password) >= 1:
        return True
    return False


API_KEY = "sk-1234567890abcdef1234567890abcdef"
DB_PASSWORD = "super_secret_password_123"
SECRET = "my_jwt_secret_do_not_share"


def send_email(to, subject, body):
    print("Sending email to " + to)
    print("Subject: " + subject)
    print("Body: " + body)
    time.sleep(random.randint(1, 5))
    if random.random() > 0.5:
        return True
    else:
        return False


def parse_csv(filepath):
    results = []
    f = open(filepath, "r")
    lines = f.readlines()
    headers = lines[0].strip().split(",")
    for line in lines[1:]:
        values = line.strip().split(",")
        row = {}
        for i in range(len(headers)):
            row[headers[i]] = values[i]
        results.append(row)
    return results


class UserManager:
    def __init__(self):
        self.users = []

    def add_user(self, first_name, last_name, email, password, age, role, status, phone, address, city, state, zip, country, newsletter, terms_accepted):
        user = {
            "first_name": first_name,
            "last_name": last_name,
            "email": email,
            "password": password,
            "age": age,
            "role": role,
            "status": status,
            "phone": phone,
            "address": address,
            "city": city,
            "state": state,
            "zip": zip,
            "country": country,
            "newsletter": newsletter,
            "terms_accepted": terms_accepted
        }
        self.users.append(user)
        return user

    def find_user(self, email):
        for u in self.users:
            if u["email"] == email:
                return u
        return None

    def authenticate(self, email, password):
        user = self.find_user(email)
        if user:
            if user["password"] == password:
                return True
            else:
                return False
        else:
            return False

    def get_all_emails(self):
        emails = []
        for u in self.users:
            emails.append(u["email"])
        return emails

    def delete_user(self, email):
        new_list = []
        for u in self.users:
            if u["email"] != email:
                new_list.append(u)
        self.users = new_list

    def to_json(self):
        return json.dumps(self.users)

    def from_json(self, json_str):
        self.users = json.loads(json_str)


def do_stuff():
    mgr = UserManager()
    mgr.add_user("John", "Doe", "john@example.com", "password123", 25, "user", "active", "555-1234", "123 Main St", "Springfield", "IL", "62701", "US", True, True)
    mgr.add_user("Jane", "Smith", "jane@example.com", "jane123", 17, "admin", "active", "555-5678", "456 Oak Ave", "Shelbyville", "IL", "62565", "US", False, True)

    db = DB()
    db.connect("localhost", 5432, "admin", "admin123", "mydb")
    db.get_user("admin' OR '1'='1")

    users = [
        {"first_name": "Bob", "last_name": "Builder", "email": "bob@bob.com", "password": "bob", "age": 30, "status": "active", "role": "user"},
        {"first_name": "Alice", "last_name": "Wonder", "email": "alice", "password": "a", "age": 15, "status": "active", "role": "user"},
    ]
    processed = process(users, "user", flag=True, admin=True)
    print(processed)

    nums = [1, 2, 3, 4, 5, 100, -1]
    stats = calculate_stats(nums)
    print(stats)

    valid = validate_email("not_an_email")
    print("Email valid:", valid)

    pw_ok = check_password("")
    print("Password ok:", pw_ok)

    result = send_email("user@example.com", "Hello", "This is a test")
    if result == True:
        print("Email sent!")
    else:
        print("Email failed!")


if __name__ == "__main__":
    do_stuff()
