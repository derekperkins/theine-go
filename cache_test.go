package theine_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/Yiling-J/theine-go"
	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	client, err := theine.New[string, string](1000)
	require.Nil(t, err)
	for i := 0; i < 20000; i++ {
		key := fmt.Sprintf("key:%d", rand.Intn(100000))
		client.Set(key, key)
	}
	time.Sleep(300 * time.Millisecond)
	require.True(t, client.Len() < 1200)
}

func TestSetParallel(t *testing.T) {
	client, err := theine.New[string, string](1000)
	require.Nil(t, err)
	var wg sync.WaitGroup
	for i := 1; i <= 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10000; i++ {
				key := fmt.Sprintf("key:%d", rand.Intn(100000))
				client.Set(key, key)
			}

		}()
	}
	wg.Wait()
	time.Sleep(300 * time.Millisecond)
	require.True(t, client.Len() < 1200)
}

func TestGetSet(t *testing.T) {
	client, err := theine.New[string, string](1000)
	require.Nil(t, err)
	for i := 0; i < 20000; i++ {
		key := fmt.Sprintf("key:%d", rand.Intn(3000))
		v, ok := client.Get(key)
		if !ok {
			client.Set(key, key)
		} else {
			require.Equal(t, key, v)
		}
	}
	time.Sleep(300 * time.Millisecond)
	require.True(t, client.Len() < 1200)
}

type foo struct {
	id int
}

func TestGeneric(t *testing.T) {
	client, err := theine.New[foo, int](100)
	require.Nil(t, err)
	key := foo{id: 1}
	client.Set(key, 1)
	v, ok := client.Get(key)
	require.True(t, ok)
	require.Equal(t, 1, v)

	v, ok = client.Get(foo{id: 1})
	require.True(t, ok)
	require.Equal(t, 1, v)

	_, ok = client.Get(foo{id: 2})
	require.False(t, ok)

	clientp, err := theine.New[*foo, int](100)
	require.Nil(t, err)
	key = foo{id: 1}
	clientp.Set(&key, 1)
	v, ok = clientp.Get(&key)
	require.True(t, ok)
	require.Equal(t, 1, v)

	_, ok = clientp.Get(&foo{id: 1})
	require.False(t, ok)
}

func TestDelete(t *testing.T) {
	client, err := theine.New[string, string](100)
	require.Nil(t, err)
	client.Set("foo", "foo")
	v, ok := client.Get("foo")
	require.True(t, ok)
	require.Equal(t, "foo", v)
	client.Delete("foo")
	_, ok = client.Get("foo")
	require.False(t, ok)

	client.SetWithTTL("foo", "foo", 10*time.Second)
	v, ok = client.Get("foo")
	require.True(t, ok)
	require.Equal(t, "foo", v)
	client.Delete("foo")
	_, ok = client.Get("foo")
	require.False(t, ok)
}

func TestGetSetParallel(t *testing.T) {
	client, err := theine.New[string, string](1000)
	require.Nil(t, err)
	var wg sync.WaitGroup
	for i := 1; i <= 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10000; i++ {
				key := fmt.Sprintf("key:%d", rand.Intn(3000))
				v, ok := client.Get(key)
				if !ok {
					client.Set(key, key)
				} else {
					require.Equal(t, key, v)
				}
			}
		}()
	}
	wg.Wait()
	time.Sleep(300 * time.Millisecond)
	require.True(t, client.Len() < 1200)
}

func TestSetWithTTL(t *testing.T) {
	client, err := theine.New[string, string](500)
	require.Nil(t, err)
	client.SetWithTTL("foo", "foo", 3600*time.Second)
	require.Equal(t, 1, client.Len())
	time.Sleep(1 * time.Second)
	client.SetWithTTL("foo", "foo", 1*time.Second)
	require.Equal(t, 1, client.Len())
	time.Sleep(2 * time.Second)
	_, ok := client.Get("foo")
	require.False(t, ok)
	require.Equal(t, 0, client.Len())
}

func TestSetWithTTLAutoExpire(t *testing.T) {
	client, err := theine.New[string, string](500)
	require.Nil(t, err)
	for i := 0; i < 30; i++ {
		key1 := fmt.Sprintf("key:%d", i)
		client.SetWithTTL(key1, key1, time.Duration(i+1)*time.Second)
		key2 := fmt.Sprintf("key:%d:2", i)
		client.SetWithTTL(key2, key2, time.Duration(i+100)*time.Second)
	}
	current := 60
	counter := 0
	for {
		time.Sleep(5 * time.Second)
		counter += 1
		require.True(t, client.Len() < current)
		current = client.Len()
		if current <= 30 {
			break
		}
	}
	require.True(t, counter < 10)
	for i := 0; i < 30; i++ {
		key := fmt.Sprintf("key:%d", i)
		_, ok := client.Get(key)
		require.False(t, ok)
	}
}

func BenchmarkGetTheineSerial(b *testing.B) {
	client, err := theine.New[string, string](100)
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(keys[rand.Intn(100000)])
	}
}

func BenchmarkGetRistrettoSerial(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(keys[rand.Intn(100000)])
	}
}

func BenchmarkGetTheineParallel(b *testing.B) {
	client, err := theine.New[string, string](100)
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			client.Get(keys[counter%100000])
			counter++
		}
	})
}

func BenchmarkGetRistrettoParallel(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			client.Get(keys[counter%100000])
			counter++
		}
	})
}

func BenchmarkSetTheineSerial(b *testing.B) {
	client, err := theine.New[string, string](100)
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Set(keys[i%100000], "")
	}
}

func BenchmarkSetRistrettoSerial(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Set(keys[i%100000], "", 1)
	}
}

func BenchmarkSetTheineParallel(b *testing.B) {
	client, err := theine.New[string, string](100)
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			client.Set(keys[counter%100000], "")
			counter++
		}
	})
}

func BenchmarkSetRistrettoParallel(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			client.Set(keys[counter%100000], "", 1)
			counter++
		}
	})
}
