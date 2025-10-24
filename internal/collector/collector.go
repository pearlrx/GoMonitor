package collector

import (
	"GoMonitor/internal/storage"
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"time"
)

type Collector struct {
	store      *storage.Postgres
	objectType string
	objectID   int
	interval   time.Duration
	cancel     context.CancelFunc
}

func New(store *storage.Postgres, objectType string, objectID int, interval time.Duration) *Collector {
	return &Collector{
		store:      store,
		objectType: objectType,
		objectID:   objectID,
		interval:   interval,
	}
}

func (c *Collector) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	ticker := time.NewTicker(c.interval)
	go func() {
		log.Printf("[collector] started, interval=%s", c.interval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Println("[collector] stopped")
				return
			case <-ticker.C:
				if err := c.collectOnce(ctx); err != nil {
					log.Printf("[collector] collect error: %v", err)
				}
			}
		}
	}()
}

func (c *Collector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Collector) collectOnce(ctx context.Context) error {
	ts := time.Now()

	// CPU percent (averaged over short interval)
	percentages, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return err
	}
	var cpuPct float64
	if len(percentages) > 0 {
		cpuPct = percentages[0]
	}

	// Memory (used in MB)
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	memoryMB := float64(vm.Used) / 1024.0 / 1024.0

	// save metrics
	if err = c.store.SaveMetric(ctx, c.objectType, c.objectID, "cpu_percent", cpuPct, ts); err != nil {
		return err
	}
	if err = c.store.SaveMetric(ctx, c.objectType, c.objectID, "memory_mb", memoryMB, ts); err != nil {
		return err
	}

	// simple log
	log.Printf("[collector] cpu=%.2f%% memory=%.1fMB server_id=%d", cpuPct, memoryMB, c.objectID)
	return nil
}
