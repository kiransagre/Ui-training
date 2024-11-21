package server

import (
	"fmt"
	"net/http"
	"training/fixlets"
)

func CreateServer() {

	http.HandleFunc("/fixlets", fixlets.GetFixlets)
	http.HandleFunc("/fixletsfromserver", fixlets.Fixletsfromserver)
	http.HandleFunc("/fixletstats", fixlets.Fixletstats)

	err := http.ListenAndServe(":8181", nil)
	if err != nil {
		fmt.Println("Unable to start server", err)
	}
}
