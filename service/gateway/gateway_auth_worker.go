package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"ergo.services/ergo/gen"
)

type authJob struct {
	connection gen.Alias
	account    string
	token      string
	frameBody  []byte
	seq        uint32
}

type authResult struct {
	job   authJob
	valid bool
}

// authWorker performs bounded blocking calls to the account verifier.
type authWorker struct {
	gen.MetaProcess
	url    string
	client *http.Client
	jobs   chan authJob
	done   chan struct{}
}

func newAuthWorker(url string) *authWorker {
	return &authWorker{
		url: url, client: &http.Client{Timeout: 3 * time.Second},
		jobs: make(chan authJob, 256), done: make(chan struct{}),
	}
}

func (w *authWorker) Init(process gen.MetaProcess) error { w.MetaProcess = process; return nil }

func (w *authWorker) Start() error {
	for {
		select {
		case job := <-w.jobs:
			_ = w.Send(w.Parent(), authResult{job: job, valid: w.verify(job)})
		case <-w.done:
			return nil
		}
	}
}

func (w *authWorker) HandleMessage(_ gen.PID, message any) error {
	job, ok := message.(authJob)
	if !ok {
		return nil
	}
	select {
	case w.jobs <- job:
		return nil
	default:
		return gen.ErrProcessMailboxFull
	}
}

func (w *authWorker) HandleCall(gen.PID, gen.Ref, any) (any, error) { return gen.ErrUnsupported, nil }
func (w *authWorker) Terminate(error)                               { close(w.done) }
func (w *authWorker) HandleInspect(gen.PID, ...string) map[string]string {
	return map[string]string{"state": "running"}
}

func (w *authWorker) verify(job authJob) bool {
	body, err := json.Marshal(map[string]string{"account": job.account, "token": job.token})
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return false
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := w.client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}
