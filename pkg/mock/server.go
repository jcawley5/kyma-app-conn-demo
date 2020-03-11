package mock

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jcawley/kyma-app-connector/pkg/utils"
)

type order struct {
	OrderCode   string  `json:"orderCode"`
	Description string  `json:"description"`
	Total       float64 `json:"total"`
}

var orders []order

//StartMockServer -
func StartMockServer(w http.ResponseWriter, r *http.Request) {

	//add some dummy data...
	orders = []order{order{"1231", "my first order", 22.2}, order{"fda2342", "my second order", 421.29}}

	go func() {

		router := mux.NewRouter().StrictSlash(true)

		log.Println("starting mock server...")
		router.HandleFunc("/orders", getOrders).Methods("GET")

		router.HandleFunc("/orders", postOrders).Methods("POST")

		// log.Fatal(http.ListenAndServeTLS(":8443", connector.GetAssetsDir()+"/kymacerts/client.crt", connector.GetAssetsDir()+"/kymacerts/private.key", router))
		log.Fatal(http.ListenAndServe(":8443", router))
	}()

	utils.ReturnSuccess("Attempting to start mock server at https://localhost:8443/orders ...", w)
}

func getOrders(w http.ResponseWriter, r *http.Request) {

	js, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func postOrders(w http.ResponseWriter, r *http.Request) {
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
	order.Total = float64(rand.Intn(max-min+1) + min)
	orders = append(orders, order)
}
