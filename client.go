package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type BidDTO struct {
	Bid string `json:"bid"`
}

const (
	fileName             = "cotacao.txt"
	clientCtxMaxDuration = 300 * time.Millisecond
)

func main() {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), clientCtxMaxDuration)
	defer clientCancel()

	bid, err := getBid(clientCtx)
	if err != nil {
		slog.Error("Error getting bid: " + err.Error())
		return
	}

	slog.Info("Current bid: " + bid)

	file, err := os.Create(fileName)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer file.Close()

	message := "Dólar: " + bid

	err = writeMessageToFile(file, message)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info("Bid saved to `" + fileName + "`")
}

func getBid(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		slog.Error("Error creating request: " + err.Error())
		return "", err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Error fetching exchange rate: " + err.Error())
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		slog.Error("Error fetching exchange rate: " + res.Status)
		return "", errors.New("server responded with status: " + res.Status)
	}

	var bidDTO BidDTO
	err = json.NewDecoder(res.Body).Decode(&bidDTO)
	if err != nil {
		slog.Error("Error decoding exchange rate: " + err.Error())
		return "", err
	}

	return bidDTO.Bid, nil
}

func writeMessageToFile(file *os.File, message string) error {
	_, err := file.WriteString(message)
	if err != nil {
		slog.Error("Error writing to file: " + err.Error())
		return err
	}

	return nil
}
