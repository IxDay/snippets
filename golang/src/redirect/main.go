package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	SSO_ENDPOINT = "https://account.booking.com/SSO/staff/auth"
)

var (
	redirect = "localhost:8090"
	timeout  = 5
	endpoint = mustParse(SSO_ENDPOINT)
)

func mustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse sso endpoint: %q", err))
	}
	return u
}

func ssoURL(redirect string) string {
	q := endpoint.Query()
	// https%3A%2F%2Fbkcloud.prod.booking.com%2Fv1%2Fredirect%3Fredirect_to%3Dhttp%3A%2F%2Flocalhost%3A8980
	q.Set("redirect_to", "https://bkcloud.prod.booking.com/v1/redirect?redirect_to=http://"+redirect)
	endpoint.RawQuery = q.Encode()
	return endpoint.String()
}

func server(addr string, handler http.Handler) Runner {
	return RunnerFunc(func(ctx context.Context) error {
		log.Println("Setting up callback server")
		lst, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		srv := &http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		log.Printf("Starting server on: %q\n", addr)
		select {
		case err := <-Async(func() error { return srv.Serve(lst) }):
			return err
		case <-ctx.Done():
			log.Println("Teardown http server")
			lst.Close()
		}
		return nil
	})

}

func interrupt() Runner {
	return InterruptCb(func() {
		fmt.Println()
		log.Println("Caught interrupt, stop serving")
	})
}

func setup() (http.Handler, Runner) {
	m := sync.Mutex{}
	m.Lock()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println("Failed to parse form from response")
			return
		}
		if token := r.Form.Get("refresh_token"); token != "" {
			log.Println("Token retrieved:", token)
		}
		w.Write([]byte{'O', 'K'})
		m.Unlock()
	})

	r := RunnerFunc(func(ctx context.Context) error {
		select {
		case <-Async(func() error {
			m.Lock()
			return nil
		}):
		case <-ctx.Done():
		}
		return nil
	})

	return h, r
}

func browser(addr string, timeout int) Runner {
	return RunnerFunc(func(ctx context.Context) error {
		for i := 0; i < timeout; i++ {
			_, err := net.Dial("tcp", addr)
			if err == nil {
				break
			}
			if i == 9 {
				return err
			}
			time.Sleep(time.Second)
		}
		url := ssoURL(addr)
		log.Printf("Redirect url is: %q\n", url)
		Open(url)
		<-ctx.Done()
		return nil
	})
}

// https://gitlab.booking.com/security/iam/wikis/
func main() {
	handler, runner := setup()
	rm := RunnerManager{}

	rm = append(rm, runner)
	rm = append(rm, browser(redirect, timeout))
	rm = append(rm, interrupt())
	rm = append(rm, server(redirect, handler))

	if err := Wait(rm); err != nil {
		log.Printf("Unexpected error: %q\n", err)
	}
}
