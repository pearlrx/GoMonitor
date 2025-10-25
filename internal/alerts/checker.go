package alerts

import (
	"log"
	"net"
	"net/http"
	"time"
)

type ServerToCheck struct {
	Name    string
	Address string
}

type Checker struct {
	Servers  []ServerToCheck
	Interval time.Duration
	Timeout  time.Duration
}

func New(servers []ServerToCheck, interval time.Duration) *Checker {
	return &Checker{
		Servers:  servers,
		Interval: interval,
		Timeout:  3 * time.Second,
	}
}

func (ch *Checker) Start() {
	status := make(map[string]bool)

	for {
		for _, srv := range ch.Servers {
			conn, err := net.DialTimeout("tcp", srv.Address, ch.Timeout)
			if err != nil {
				if !status[srv.Name] {
					log.Printf("[ALERT] Сервер %s (%s) недоступен (TCP): %v", srv.Name, srv.Address, err)
					status[srv.Name] = true
				} else {
					log.Printf("[DOWN] Сервер %s (%s) всё ещё недоступен...", srv.Name, srv.Address)
				}
				continue
			}
			conn.Close()

			client := http.Client{
				Timeout: ch.Timeout,
			}
			resp, err := client.Get("http://" + srv.Address)
			if err != nil {
				if !status[srv.Name] {
					log.Printf("[ALERT] Сервер %s (%s) отвечает по TCP, но не по HTTP: %v", srv.Name, srv.Address, err)
					status[srv.Name] = true
				} else {
					log.Printf("[DOWN] Сервер %s (%s) всё ещё не отвечает по HTTP...", srv.Name, srv.Address)
				}
				continue
			}
			resp.Body.Close()

			if status[srv.Name] {
				log.Printf("[RECOVERED] Сервер %s (%s) снова доступен (HTTP %d)", srv.Name, srv.Address, resp.StatusCode)
				status[srv.Name] = false
			} else {
				log.Printf("[OK] Сервер %s (%s) доступен (HTTP %d)", srv.Name, srv.Address, resp.StatusCode)
			}
		}

		time.Sleep(ch.Interval)
	}
}
