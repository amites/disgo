package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dispatchlabs/dapos"
	daposCore "github.com/dispatchlabs/dapos/core"

	httpService "github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/disgover"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/dispatchlabs/commons/utils"
	"time"
)

// Api
type Api struct {
	router *mux.Router
}

// NewApi
func NewApi() *Api {
	this := Api{httpService.GetHttpRouter()}
	this.router.HandleFunc("/v1/ping", this.pingPongHandler).Methods("POST")
	this.router.HandleFunc("/v1/balance/{address}", this.retrieveBalanceHandler).Methods("GET")
	this.router.HandleFunc("/v1/sync_transactions", this.syncTransactionsHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions/{address}", this.retrieveTransactionsHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions", this.createTransactionHandler).Methods("POST")
	this.router.HandleFunc("/v1/test_transaction", this.createTestTransactionHandler).Methods("POST")
	return &this
}

// retrieveBalanceHandler
func (this *Api) retrieveBalanceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	balance, error := dapos.GetDAPoS().GetBalance(vars["address"])
	if error != nil {
		responseWriter.Write([]byte(`{"status":"INTERNAL_SERVER_ERROR"}`))
		return
	}
	bytes, error := json.Marshal(struct {
		Status  string `json:"status,omitempty"`
		Balance int64  `json:"balance,omitempty"`
	}{
		Status:  "OK",
		Balance: balance,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}
	responseWriter.Write(bytes)
}

// createTransactionHandler
func (this *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("unable to read HTTP body of request [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	// Unmarshal transaction?
	transaction := &daposCore.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	dapos.GetDAPoS().ProcessTx(transaction)
	responseWriter.Write([]byte(`{"status":"OK"}`))
}

// createTransactionHandler
func (this *Api) createTestTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("unable to read HTTP body of request [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	var jsonMap map[string]interface{}
	err := json.Unmarshal(body, &jsonMap)
	if err != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	transaction := daposCore.NewTransaction(
		jsonMap["privateKey"].(string),
		0,
		jsonMap["from"].(string),
		jsonMap["to"].(string),
		int64(jsonMap["value"].(float64)),
		time.Now(),
	)

	dapos.GetDAPoS().ProcessTx(transaction)
	responseWriter.Write([]byte(`{"status":"OK"}`))
}

func (this *Api) syncTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	dapos.GetDAPoS().SynchronizeTransactions()
}

// retrieveTransactionsHandler
func (this *Api) retrieveTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	transactions := dapos.GetDAPoS().GetTransactions(vars["address"])
	bytes, error := json.Marshal(struct {
		Status       string                  `json:"status,omitempty"`
		Transactions []daposCore.Transaction `json:"transactions,omitempty"`
	}{
		Status:       "OK",
		Transactions: transactions,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}
	responseWriter.Write(bytes)
}

func (this *Api) pingPongHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	fmt.Println(string(body))

	responseWriter.Write([]byte(fmt.Sprintf(
		"PONG-From: %s @ %s:%d",
		disgover.GetDisgover().ThisContact.Address,
		disgover.GetDisgover().ThisContact.Endpoint.Host,
		disgover.GetDisgover().ThisContact.Endpoint.Port,
	)))
}
