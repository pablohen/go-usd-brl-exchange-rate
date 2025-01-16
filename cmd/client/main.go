package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

type BidDTO struct {
	Bid string `json:"bid"`
}

const fileName = "cotacao.txt"

func main() {
	bid, err := getBid()
	if err != nil {
		slog.Error("Error getting bid", "error", err.Error())
		panic(err)
	}

	slog.Info("Current bid: " + bid)

	file, err := os.Create(fileName)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	message := "Dólar: " + bid

	err = writeMessageToFile(file, message)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	slog.Info("Bid saved to `" + fileName + "`")
}

func getBid() (string, error) {
	res, err := http.Get("http://localhost:8080/cotacao")
	if err != nil {
		slog.Error("Error fetching exchange rate", "error", err.Error())
		return "", err
	}
	defer res.Body.Close()

	var bidDTO BidDTO
	err = json.NewDecoder(res.Body).Decode(&bidDTO)
	if err != nil {
		slog.Error("Error decoding exchange rate", "error", err.Error())
		return "", err
	}

	return bidDTO.Bid, nil
}

func writeMessageToFile(file *os.File, message string) error {
	_, err := file.WriteString(message)
	if err != nil {
		slog.Error("Error writing to file", "error", err.Error())
		return err
	}

	return nil
}
