package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	_ "github.com/mattn/go-sqlite3"
)

const (
	apiBaseURL           = "https://economia.awesomeapi.com.br"
	dataSourceName       = "bids.db"
	createTableStatement = `CREATE TABLE IF NOT EXISTS bids (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL
	)`
	insertNewBidStatement = "INSERT INTO bids (bid) VALUES (?)"
	dbCtxMaxDuration      = 10 * time.Millisecond
	apiCtxMaxDuration     = 200 * time.Millisecond
)

func main() {
	db, err := getDb(dataSourceName)
	if err != nil {
		slog.Error("Failed to get database "+dataSourceName, "error", err.Error())
		panic(err)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbCtxMaxDuration)
	defer dbCancel()

	apiCtx, apiCancel := context.WithTimeout(context.Background(), apiCtxMaxDuration)
	defer apiCancel()

	h := &ExchangeHandler{
		db:     db,
		dbCtx:  &dbCtx,
		apiCtx: &apiCtx,
	}

	s := fuego.NewServer(
		fuego.WithAddr("localhost:8080"),
	)

	fuego.GetStd(s, "/cotacao", h.getBid,
		option.Summary("Get bid"),
		option.Description("Get current bid from USD to BRL"),
		option.AddResponse(200, "", fuego.Response{
			Type: GetBidOutput{},
		}),
	)

	s.Run()
}

func getDb(datasource string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./"+datasource)
	if err != nil {
		slog.Error("Failed to open database: " + err.Error())
		return nil, err
	}

	_, err = db.Exec(createTableStatement)
	if err != nil {
		slog.Error("Failed to create table: " + err.Error())
		return nil, err
	}

	return db, nil
}

func (h *ExchangeHandler) getBid(w http.ResponseWriter, r *http.Request) {
	rate, err := getCurrentExchangeRate(*h.apiCtx)
	if err != nil {
		slog.Error("Error fetching exchange rate: " + err.Error())
		http.Error(w, "Error fetching exchange rate", http.StatusInternalServerError)
		return
	}
	bid := rate.UsdBrl.Bid
	slog.Info("Current bid: " + bid)

	err = saveBid(*h.dbCtx, h.db, bid)
	if err != nil {
		slog.Error("Error saving exchange rate: " + err.Error())
		http.Error(w, "Error saving exchange rate", http.StatusInternalServerError)
		return
	}
	slog.Info("Exchange rate saved successfully")

	w.Header().Set("Content-Type", "application/json")

	response := GetBidOutput{Bid: bid}
	json.NewEncoder(w).Encode(response)
}

func getCurrentExchangeRate(ctx context.Context) (*CurrentExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiBaseURL+"/json/last/USD-BRL", nil)
	if err != nil {
		slog.Error("Error creating request: " + err.Error())
		return &CurrentExchangeRate{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error doing request: " + err.Error())
		return &CurrentExchangeRate{}, err
	}
	defer resp.Body.Close()

	var response CurrentExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("Error decoding response: " + err.Error())
		return &CurrentExchangeRate{}, err
	}

	return &response, nil
}

func saveBid(ctx context.Context, db *sql.DB, bid string) error {
	stmt, err := db.PrepareContext(ctx, insertNewBidStatement)
	if err != nil {
		slog.Error("Error preparing statement: " + err.Error())
		return err
	}

	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
		slog.Error("Error executing statement: " + err.Error())
		return err
	}

	return nil
}

type ExchangeHandler struct {
	db     *sql.DB
	dbCtx  *context.Context
	apiCtx *context.Context
}

type CurrentExchangeRate struct {
	UsdBrl UsdBrl `json:"USDBRL"`
}

type UsdBrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type GetBidOutput struct {
	Bid string `json:"bid"`
}
