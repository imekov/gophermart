package handlers

import "net/http"

func (h Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, err := w.Write([]byte("Защищенная ручка"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) PostUserOrders(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {

}
func (h Handler) PostBalanceWithdraw(w http.ResponseWriter, r *http.Request) {

}
