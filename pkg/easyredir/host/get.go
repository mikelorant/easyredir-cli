package host

import (
	"fmt"
	"net/http"

	"github.com/mikelorant/easyredir-cli/pkg/jsonutil"
)

func GetHost(cl ClientAPI, id string) (h Host, err error) {
	pathQuery := buildGetHost(id)
	reader, err := cl.SendRequest(pathQuery, http.MethodGet, nil)
	if err != nil {
		return h, fmt.Errorf("unable to send request: %w", err)
	}

	if err := jsonutil.DecodeJSON(reader, &h); err != nil {
		return h, fmt.Errorf("unable to get json: %w", err)
	}

	if ok := (h.Data.ID == id); !ok {
		return h, fmt.Errorf("received incorrect host: %v", h.Data.ID)
	}

	return h, nil
}

func buildGetHost(id string) string {
	return fmt.Sprintf("/hosts/%v", id)
}
