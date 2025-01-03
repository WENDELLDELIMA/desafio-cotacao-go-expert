package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func getCotacao() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição ao servidor: %v", err)
	}
	defer resp.Body.Close()

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return data["bid"], nil
}

func saveToFile(cotacao string) error {
	fileContent := fmt.Sprintf("Dólar: %s", cotacao)
	err := ioutil.WriteFile("cotacao.txt", []byte(fileContent), 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar no arquivo: %v", err)
	}
	return nil
}

func main() {
	cotacao, err := getCotacao()
	if err != nil {
		log.Fatalf("Erro ao obter cotação: %v", err)
	}

	fmt.Println("Cotação recebida:", cotacao)

	if err := saveToFile(cotacao); err != nil {
		log.Fatalf("Erro ao salvar cotação no arquivo: %v", err)
	}

	fmt.Println("Cotação salva com sucesso em cotacao.txt")
}
