package account

import (
	"sync"
	"testing"

	"ds2api/internal/config"
)

func newPoolForTest(t *testing.T, maxInflight string) *Pool {
	t.Helper()
	t.Setenv("DS2API_ACCOUNT_MAX_INFLIGHT", maxInflight)
	t.Setenv("DS2API_CONFIG_JSON", `{
		"keys":["k1"],
		"accounts":[
			{"email":"acc1@example.com","token":"token1"},
			{"email":"acc2@example.com","token":"token2"}
		]
	}`)
	store := config.LoadStore()
	return NewPool(store)
}

func TestPoolRoundRobinWithConcurrentSlots(t *testing.T) {
	pool := newPoolForTest(t, "2")

	order := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		acc, ok := pool.Acquire("", nil)
		if !ok {
			t.Fatalf("expected acquire success at step %d", i+1)
		}
		order = append(order, acc.Identifier())
	}
	want := []string{"acc1@example.com", "acc2@example.com", "acc1@example.com", "acc2@example.com"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected order at %d: got %q want %q (full=%v)", i, order[i], want[i], order)
		}
	}

	if _, ok := pool.Acquire("", nil); ok {
		t.Fatalf("expected acquire to fail when all inflight slots are occupied")
	}

	pool.Release("acc1@example.com")
	acc, ok := pool.Acquire("", nil)
	if !ok || acc.Identifier() != "acc1@example.com" {
		t.Fatalf("expected reacquire acc1 after releasing one slot, got ok=%v id=%q", ok, acc.Identifier())
	}
}

func TestPoolTargetAccountInflightLimit(t *testing.T) {
	pool := newPoolForTest(t, "2")

	for i := 0; i < 2; i++ {
		if _, ok := pool.Acquire("acc1@example.com", nil); !ok {
			t.Fatalf("expected target acquire success at step %d", i+1)
		}
	}
	if _, ok := pool.Acquire("acc1@example.com", nil); ok {
		t.Fatalf("expected third acquire on same target to fail due to inflight limit")
	}
}

func TestPoolConcurrentAcquireDistribution(t *testing.T) {
	pool := newPoolForTest(t, "2")

	start := make(chan struct{})
	results := make(chan string, 6)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			acc, ok := pool.Acquire("", nil)
			if !ok {
				results <- "FAIL"
				return
			}
			results <- acc.Identifier()
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	success := 0
	fail := 0
	perAccount := map[string]int{}
	for id := range results {
		if id == "FAIL" {
			fail++
			continue
		}
		success++
		perAccount[id]++
	}
	if success != 4 || fail != 2 {
		t.Fatalf("unexpected concurrent acquire result: success=%d fail=%d perAccount=%v", success, fail, perAccount)
	}
	for id, n := range perAccount {
		if n > 2 {
			t.Fatalf("account %s exceeded inflight limit: %d", id, n)
		}
	}
}

func TestPoolStatusRecommendedConcurrencyDefault(t *testing.T) {
	pool := newPoolForTest(t, "")
	status := pool.Status()

	if got, ok := status["max_inflight_per_account"].(int); !ok || got != 2 {
		t.Fatalf("unexpected max_inflight_per_account: %#v", status["max_inflight_per_account"])
	}
	if got, ok := status["recommended_concurrency"].(int); !ok || got != 4 {
		t.Fatalf("unexpected recommended_concurrency: %#v", status["recommended_concurrency"])
	}
}

func TestPoolStatusRecommendedConcurrencyRespectsOverride(t *testing.T) {
	pool := newPoolForTest(t, "3")
	status := pool.Status()

	if got, ok := status["max_inflight_per_account"].(int); !ok || got != 3 {
		t.Fatalf("unexpected max_inflight_per_account: %#v", status["max_inflight_per_account"])
	}
	if got, ok := status["recommended_concurrency"].(int); !ok || got != 6 {
		t.Fatalf("unexpected recommended_concurrency: %#v", status["recommended_concurrency"])
	}
}

func TestPoolAccountConcurrencyAliasEnv(t *testing.T) {
	t.Setenv("DS2API_ACCOUNT_MAX_INFLIGHT", "")
	t.Setenv("DS2API_ACCOUNT_CONCURRENCY", "4")
	t.Setenv("DS2API_CONFIG_JSON", `{
		"keys":["k1"],
		"accounts":[
			{"email":"acc1@example.com","token":"token1"},
			{"email":"acc2@example.com","token":"token2"}
		]
	}`)

	pool := NewPool(config.LoadStore())
	status := pool.Status()
	if got, ok := status["max_inflight_per_account"].(int); !ok || got != 4 {
		t.Fatalf("unexpected max_inflight_per_account: %#v", status["max_inflight_per_account"])
	}
	if got, ok := status["recommended_concurrency"].(int); !ok || got != 8 {
		t.Fatalf("unexpected recommended_concurrency: %#v", status["recommended_concurrency"])
	}
}

func TestPoolSupportsTokenOnlyAccount(t *testing.T) {
	t.Setenv("DS2API_ACCOUNT_MAX_INFLIGHT", "1")
	t.Setenv("DS2API_CONFIG_JSON", `{
		"keys":["k1"],
		"accounts":[{"token":"token-only-account"}]
	}`)

	pool := NewPool(config.LoadStore())
	status := pool.Status()
	if got, ok := status["total"].(int); !ok || got != 1 {
		t.Fatalf("unexpected total in pool status: %#v", status["total"])
	}
	if got, ok := status["available"].(int); !ok || got != 1 {
		t.Fatalf("unexpected available in pool status: %#v", status["available"])
	}

	acc, ok := pool.Acquire("", nil)
	if !ok {
		t.Fatalf("expected acquire success for token-only account")
	}
	if acc.Token != "token-only-account" {
		t.Fatalf("unexpected token on acquired account: %q", acc.Token)
	}
}
