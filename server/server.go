package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

type Cotacao struct {
	USD struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func requestCotacao(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return "", err
	}

	return cotacao.USD.Bid, nil
}

func salvarCotacao(ctx context.Context, db *sql.DB, bid string) error {
	query := "INSERT INTO cotacoes (bid, timestamp) VALUES (?, ?)"
	statement, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, bid, time.Now())
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {

	bid, err := requestCotacao(r.Context())
	if err != nil {
		http.Error(w, "Falha ao realizar request da cotação.", http.StatusInternalServerError)
		return
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer dbCancel()

	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		http.Error(w, "Erro ao conectar no banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := salvarCotacao(dbCtx, db, bid); err != nil {
		http.Error(w, "Erro ao salvar cotação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}
