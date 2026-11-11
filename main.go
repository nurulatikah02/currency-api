package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ConvertRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type RateResponse struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

type ConvertResult struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Result float64 `json:"result"`
	Date   string  `json:"date"`
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Gunakan metode POST",
		})
		return
	}

	var input ConvertRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, `{"error":"JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	input.From = strings.ToUpper(strings.TrimSpace(input.From))
	input.To = strings.ToUpper(strings.TrimSpace(input.To))

	if input.From == "" || input.To == "" || input.Amount <= 0 {
		http.Error(w, `{"error":"Data mata uang atau jumlah tidak valid"}`, http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf(
		"https://api.frankfurter.dev/v2/rate/%s/%s",
		input.From,
		input.To,
	)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, `{"error":"Gagal menghubungi API kurs"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, `{"error":"Mata uang tidak ditemukan atau kurs tidak tersedia"}`, http.StatusBadGateway)
		return
	}

	var rate RateResponse

	err = json.NewDecoder(resp.Body).Decode(&rate)
	if err != nil {
		http.Error(w, `{"error":"Gagal membaca data kurs"}`, http.StatusBadGateway)
		return
	}

	result := ConvertResult{
		From:   input.From,
		To:     input.To,
		Amount: input.Amount,
		Rate:   rate.Rate,
		Result: input.Amount * rate.Rate,
		Date:   rate.Date,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/api/convert", convertHandler)

	fmt.Println("Currency API berjalan di http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
