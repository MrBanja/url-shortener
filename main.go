package main

import (
	"context"
	"log"

	"github.com/caarlos0/env/v10"

	"github.com/mrbanja/url-shortener/app"
)

func main() {
	var o app.Options
	if err := env.Parse(&o); err != nil {
		log.Fatalf("parse options: %v", err)
	}

	ctx := context.Background()
	if err := app.Run(ctx, &o); err != nil {
		log.Panicf("run app: %v", err)
	}
}
