package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	url = "http://localhost:8080/cotacao"
)

type Cotacao struct {
	Valor string `json:"Dólar"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	bid, err := solicitarCotacao(ctx)
	if err != nil {
		log.Println("Error ao solicitar cotação:", err)
		return
	}

	if _, err := salvarArquivo(bid); err != nil {
		log.Println("Erro ao salvar cotacao no arquivo:", err)
		return
	}
}

func solicitarCotacao(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["bid"], nil
}

func salvarArquivo(bid string) (len int, err error) {
	conteudo := fmt.Sprintf("Dólar: %s", bid)
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write([]byte(conteudo))
}
