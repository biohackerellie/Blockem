package api

import (
	"Blockem/internal/server"
	"fmt"
)

func api() {

	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
