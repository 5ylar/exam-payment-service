package main

import (
	"database/sql"
	paymentServer "exam-payment-service/api/payment"
	"exam-payment-service/internal/payment"
	"exam-payment-service/pkg/omiseprovider"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/omise/omise-go"
)

func main() {
	var (
		port           string = os.Getenv("PORT")
		omisePublicKey string = os.Getenv("OMISE_PUBLIC_KEY")
		omiseSecretKey string = os.Getenv("OMISE_SECRET_KEY")
	)

	// Omise client
	oc, err := omise.NewClient(omisePublicKey, omiseSecretKey)
	if err != nil {
		panic(err)
	}

	// Omise provider
	op := omiseprovider.New(oc)

	// Database
	db, err := sql.Open("sqlite3", "./payment.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Prepare table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			charge_id 	varchar(100) NOT NULL PRIMARY KEY,
			source_id 	varchar(100),
			txn_id 	varchar(100),
			status 		varchar(20),
			UNIQUE (charge_id)
		)
	`,
	)
	if err != nil {
		panic(err)
	}

	// Payment
	p := payment.New(op, db)

	// Payment server
	paymentServer.Start(":"+port, p)
}
