package projector

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Error      string      `json:"error"`
	Subject    interface{} `json:"subject,omitempty"`
	Collection interface{} `json:"collection,omitempty"`
}

func (self *Response) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.MarshalIndent(self, "", "  ")

	if self.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	fmt.Fprintf(w, "%s\n", data)
}
