package main

import (
	"log"
	"os"
)

func main() {
	for _, arg := range os.Args {
		switch arg {
		case ucCreateTransactionRaceCondition:
			createTransactionRaceCondition()
			return
		case ucRevertTransactionRaceCondition:
			revertTransactionRaceCondition()
			return
		case ucCreateHoldConcurrency:
			createHoldConcurrency()
			return
		default:
			log.Print("case not implemented")
		}
	}
}
