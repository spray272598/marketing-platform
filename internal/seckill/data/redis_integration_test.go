package data

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisLuaStockDecrement(t *testing.T) {
	// Skip if no Redis available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// Ping Redis
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	activityID := "test_stock_" + time.Now().Format("20060102150405")
	key := fmt.Sprintf("seckill:stock:%s", activityID)

	// Set initial stock
	initialStock := int32(100)
	err := client.Set(ctx, key, initialStock, 10*time.Minute).Err()
	if err != nil {
		t.Fatalf("failed to set stock: %v", err)
	}

	// Lua script for atomic stock deduction
	script := redis.NewScript(`
		local stock = tonumber(redis.call('GET', KEYS[1]))
		if stock == nil then
			return 0
		end
		if stock <= 0 then
			return 0
		end
		redis.call('DECR', KEYS[1])
		return 1
	`)

	// Concurrent requests
	concurrentCount := 200
	var successCount int64
	var failCount int64

	var wg sync.WaitGroup
	wg.Add(concurrentCount)

	start := time.Now()

	for i := 0; i < concurrentCount; i++ {
		go func(userID int) {
			defer wg.Done()

			result, err := script.Run(ctx, client, []string{key}).Int()
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}

			if result == 1 {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Get final stock
	finalStock, _ := client.Get(ctx, key).Int()

	t.Logf("=== Redis Lua Stock Decrement Test ===")
	t.Logf("Initial stock: %d", initialStock)
	t.Logf("Concurrent requests: %d", concurrentCount)
	t.Logf("Success count: %d", successCount)
	t.Logf("Fail count: %d", failCount)
	t.Logf("Final stock: %d", finalStock)
	t.Logf("Duration: %v", duration)
	t.Logf("QPS: %.2f", float64(concurrentCount)/duration.Seconds())

	// Verify
	if int32(finalStock) != initialStock-int32(successCount) {
		t.Errorf("stock mismatch: expected %d, got %d", initialStock-int32(successCount), finalStock)
	}

	if successCount+failCount != int64(concurrentCount) {
		t.Errorf("total count mismatch: expected %d, got %d", concurrentCount, successCount+failCount)
	}

	// Stock should never be negative
	if finalStock < 0 {
		t.Errorf("stock is negative: %d", finalStock)
	}
}

func TestRedisDistributedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	lockKey := "test_lock_" + time.Now().Format("20060102150405")

	// Test lock acquisition
	ok, err := client.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	if !ok {
		t.Fatal("failed to acquire lock: already locked")
	}

	// Test lock conflict
	ok2, err := client.SetNX(ctx, lockKey, "2", 10*time.Second).Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok2 {
		t.Fatal("should not acquire lock: already locked")
	}

	// Release lock
	err = client.Del(ctx, lockKey).Err()
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Test lock after release
	ok3, err := client.SetNX(ctx, lockKey, "3", 10*time.Second).Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok3 {
		t.Fatal("should acquire lock after release")
	}

	t.Logf("Distributed lock test passed")

	// Cleanup
	client.Del(ctx, lockKey)
}

func TestRedisConcurrentLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	lockKey := "test_concurrent_lock_" + time.Now().Format("20060102150405")

	// Only one should acquire the lock
	concurrentCount := 50
	var successCount int64

	var wg sync.WaitGroup
	wg.Add(concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func() {
			defer wg.Done()

			ok, _ := client.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
			if ok {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	t.Logf("=== Concurrent Lock Test ===")
	t.Logf("Concurrent attempts: %d", concurrentCount)
	t.Logf("Successful locks: %d", successCount)

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful lock, got %d", successCount)
	}

	// Cleanup
	client.Del(ctx, lockKey)
}
