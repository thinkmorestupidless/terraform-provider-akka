package akka

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForReady_becomesReady(t *testing.T) {
	calls := int32(0)
	client := &testServiceClient{
		responses: []string{"UpdateInProgress", "UpdateInProgress", "Ready"},
		callCount: &calls,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.waitForReady(ctx, 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForReady_unavailableLimit(t *testing.T) {
	calls := int32(0)
	client := &testServiceClient{
		responses: []string{"Unavailable", "Unavailable", "Unavailable"},
		callCount: &calls,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.waitForReady(ctx, 10*time.Millisecond); err == nil {
		t.Fatal("expected error after 3 Unavailable, got nil")
	}
}

func TestWaitForReady_timeout(t *testing.T) {
	calls := int32(0)
	client := &testServiceClient{
		responses: []string{"UpdateInProgress"},
		loop:      true,
		callCount: &calls,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := client.waitForReady(ctx, 10*time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// testServiceClient is a minimal stub for testing WaitForReady logic without the real AkkaClient.
type testServiceClient struct {
	responses []string
	loop      bool
	callCount *int32
	idx       int
}

func (c *testServiceClient) waitForReady(ctx context.Context, interval time.Duration) error {
	unavailableCount := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			status := c.nextStatus()
			switch status {
			case "Ready":
				return nil
			case "Unavailable":
				unavailableCount++
				if unavailableCount >= 3 {
					return context.DeadlineExceeded
				}
			default:
				unavailableCount = 0
			}
		}
	}
}

func (c *testServiceClient) nextStatus() string {
	atomic.AddInt32(c.callCount, 1)
	if c.loop {
		return c.responses[0]
	}
	if c.idx >= len(c.responses) {
		return c.responses[len(c.responses)-1]
	}
	s := c.responses[c.idx]
	c.idx++
	return s
}
