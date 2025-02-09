package cache

import (
	"sync"
	"time"
)

type InputCache struct { // структура со следующими аргументами
	Data     interface{} // входные данные (можем хранить любой тим данных)
	TimeLife int64       // время жизни наших данных
}

type Cache struct { // эта структура хранит наш кэш
	safe map[string]InputCache // тут записывается название ключа, а значение - данные и время жизни
	mu   sync.RWMutex
} // мапа - не потокобезопасна, поэтому используем RWMutex
/*
RWMutex - позволяет читать всем горутинам, но делать запись
только одной (RLock - читать одновременно, Lock - записывать)
*/

func NewCache() *Cache { // создаем пустую мапу для сейфа с указателем, чтобы передавать ссылку

	return &Cache{safe: make(map[string]InputCache)}
}

// добавляем новый элемент в кэш
func (c *Cache) Set(key string, data any, ttl time.Duration) {
	c.mu.Lock()         // блокируем наш кэш для записи
	defer c.mu.Unlock() // defer сработает только после завершения функции и откроет кэш

	c.safe[key] = InputCache{ // присваиваем ключу из мапы safe входные значения.
		Data:     data,
		TimeLife: time.Now().Add(ttl).UnixNano(),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock() // блокируем кэш для чтения
	defer c.mu.RUnlock()

	safe, found := c.safe[key]                           // присваеваем значение из ключа и проверяем, если там что-то
	if !found || time.Now().UnixNano() > safe.TimeLife { // или нет значения или время вышло
		return nil, false
	}
	return safe.Data, true // получаем нужные нам данные из кэша
}

func (c *Cache) Delete(key string) { // удаляем ключ из кэша
	c.mu.Lock() // блокируем наш кэш для удаления
	defer c.mu.Unlock()
	delete(c.safe, key)
}

func (c *Cache) WorkerPool(interval time.Duration, LotWorker int) { // создаем воркеров и проверку кэша

	jobs := make(chan string, LotWorker) // делаем канал, который передает кэш на удаление воркерам
	var wg sync.WaitGroup                // механизм ожидания завершения горутины

	for i := 0; i < LotWorker; i++ { // создаем воркеров по лимит не привысим
		wg.Add(1) // увеличиваем кол-во активных рабочих
		go func() {
			defer wg.Done()
			for key := range jobs {
				c.Delete(key)
			}
		}()
	}
	go func() {
		for {
			time.Sleep(interval)            // блокируем проверку, чтобы не грузить проц
			c.mu.Lock()                     // блокируем запись в мапу
			now := time.Now().UnixNano()    // узнаем время сейчас
			for key, safe := range c.safe { // проходимся по кэшу
				if safe.TimeLife < now { // сравниваем истек ли сорк
					jobs <- key // если да, то оптравляем в канал к воркерам
				}
			}
			c.mu.Unlock()
		}

	}()
	// lol
}
