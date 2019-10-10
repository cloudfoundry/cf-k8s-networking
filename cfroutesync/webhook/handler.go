package webhook

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cf-networking-helpers/marshal"
)

//go:generate counterfeiter -o fakes/syncer.go --fake-name Syncer . syncer
type syncer interface {
	Sync(syncRequest SyncRequest) (*SyncResponse, error)
}

type SyncHandler struct {
	Marshaler   marshal.Marshaler
	Unmarshaler marshal.Unmarshaler
	Syncer      syncer
}

// ServeHTTP serves the /sync webhook to metacontroller
func (r *SyncHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		r.respondWithCode(http.StatusInternalServerError, rw, "failed to read request")
		return
	}

	syncRequest := &SyncRequest{}
	err = r.Unmarshaler.Unmarshal(bodyBytes, syncRequest)
	if err != nil {
		r.respondWithCode(http.StatusBadRequest, rw, "failed to unmarshal request")
		return
	}

	response, err := r.Syncer.Sync(*syncRequest)
	if err != nil {
		if err == UninitializedError {
			r.respondWithCode(http.StatusInternalServerError, rw, err.Error())
		} else {
			r.respondWithCode(http.StatusInternalServerError, rw, "Internal Server Error")
		}
		return
	}
	bytes, err := r.Marshaler.Marshal(response)
	if err != nil {
		r.respondWithCode(http.StatusInternalServerError, rw, "failed to marshal response")
		return
	}
	rw.Write(bytes)
}

func (r *SyncHandler) respondWithCode(statusCode int, w http.ResponseWriter, description string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, description)))
}
