//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/dpdlv-sdk/dpdlv"
)

func main() {
	client := dpdlv.New(dpdlv.Config{
		Username:    os.Getenv("DPD_USERNAME"),
		Password:    os.Getenv("DPD_PASSWORD"),
		Environment: dpdlv.EnvSandbox,
	})
	api, auth, err := client.API(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("base", dpdlv.BaseURL(client.Environment()), "api=", api != nil, "auth=", auth != nil)
}
