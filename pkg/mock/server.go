package mock

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
)

type order struct {
	OrderCode   string  `json:"orderCode"`
	Description string  `json:"description"`
	Total       float64 `json:"total"`
}

var orders []order

func init() {
	//add some dummy data...
	orders = []order{order{"1231", "my first order", 22.2}, order{"fda2342", "my second order", 421.29}}
}

//GetOrders -
func GetOrders(w http.ResponseWriter, r *http.Request) {

	js, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//GetOrder -
func GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, "ID: %v\n", vars["id"])

	// sort.SearchStrings(orders, "A")

	id := vars["id"]
	idx := sort.Search(len(orders), func(i int) bool {
		return string(orders[i].OrderCode) >= id
	})

	w.Header().Set("Content-Type", "application/json")
	if orders[idx].OrderCode == id {
		js, err := json.Marshal(orders[idx])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	} else {
		w.Write([]byte("{}"))
	}
}

//PostOrders -
func PostOrders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	orderData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var order order
	err = json.Unmarshal(orderData, &order)

	orders = append(orders, order)

	js, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//AddOrderFromEvent -
func AddOrderFromEvent(orderCode string) {

	rand.Seed(time.Now().UnixNano())
	min := 10
	max := 400

	var order order
	order.OrderCode = orderCode
	order.Description = "Order created from event"
	rand.Float64()
	order.Total = float64(rand.Intn(max-min+1)+min) + .99
	orders = append(orders, order)
}
