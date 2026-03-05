package pkg

import (
	"log"
	"fmt"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger(debug bool) {
	var err error

	if debug {
		Logger, err = zap.NewDevelopment()
		fmt.Println("✓ Setting development logger done")
	} else {
		Logger, err = zap.NewProduction()
		fmt.Println("✓ Setting production logger done")
	}
	if err != nil {
		log.Fatalf("init logger error: %v", err)
	}
}
