package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db := setupDB()
	defer db.Close()

	http.HandleFunc("/cotacao", cotacaoHandler(db))
	fmt.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func setupDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatal("Erro ao abrir banco de dados:", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Erro ao criar tabela:", err)
	}

	return db
}
func cotacaoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			http.Error(w, "Erro ao criar requisição", http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Erro na requisição à API:", err)
			http.Error(w, "Timeout na requisição", http.StatusGatewayTimeout)
			return
		}
		defer resp.Body.Close()

		var data map[string]map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			http.Error(w, "Erro ao decodificar resposta", http.StatusInternalServerError)
			return
		}

		bid := data["USDBRL"]["bid"]

		// Timeout para salvar no banco
		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		_, err = db.ExecContext(ctxDB, "INSERT INTO cotacoes (valor) VALUES (?)", bid)
		if err != nil {
			log.Println("Erro ao salvar cotação:", err)
			http.Error(w, "Timeout ao salvar no banco", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": bid})
	}
}
