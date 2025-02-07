package cache

import (
	"testing"
	"time"
)

// Тест Set и Get
func TestCache_SetAndGet(t *testing.T) {
	c := NewCache()
	c.Set("username", "Dima", 2*time.Second)

	safe, found := c.Get("username")
	if !found || safe != "Dima" {
		t.Errorf("Ожидалось 'Dima', получено %v", safe)
	}
}

// Тест истечения TTL
func TestCache_TTLExpiration(t *testing.T) {
	c := NewCache()
	c.Set("username", "Dima", 1*time.Second)
	time.Sleep(2 * time.Second) // Ждём истечения TTL

	safe, found := c.Get("username")
	if found || safe != nil {
		t.Errorf("Ожидалось nil, ключ должен был истечь, но получено %v", safe)
	}
}

// Тест удаления ключа
func TestCache_Delete(t *testing.T) {
	c := NewCache()
	c.Set("username", "Dima", 10*time.Second)
	c.Delete("username")

	safe, found := c.Get("username")
	if found || safe != nil {
		t.Errorf("Ожидалось nil, ключ должен был быть удалён, но получено %v", safe)
	}
}

// Тест автоматической очистки
func TestCache_AutoEviction(t *testing.T) {
	c := NewCache()
	c.Set("temp", "data", 1*time.Second)
	c.WorkerPool(500*time.Millisecond, 2)
	time.Sleep(2 * time.Second)

	safe, found := c.Get("temp")
	if found || safe != nil {
		t.Errorf("Ожидалось nil, ключ должен был быть автоматически удалён, но получено %v", safe)
	}
}
