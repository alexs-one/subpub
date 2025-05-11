
# SubPub service

1. Имплементирована библиотека subpub из задания:

## Метод Subscribe

Добавляет нового подписчика в inmemory пул. У подпсчика инициализируются три поля: callback функция, буферизированный канал с сообщениями, канал для остановки чтения клиентом. 

Чтение сообщений и вызов callback функции происходит в запускаемой горутине. Горутина отслеживает завершение чтения клиентом, а также закрытие канала subpub при завершении работы сервиса. 

Возвращает subscription у которого имплементирован единственный метод Unsubscribe()

```golang
func (sp *subPub) Subscribe(subject string, cb MessageHandler) Subscription {
	sub := &subscriber{
		callback: cb,
		ch:       make(chan interface{}, 10),
		stop:     make(chan struct{}),
	}

	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.subscribers[subject] = append(sp.subscribers[subject], sub)
	sp.shutdownWg.Add(1)
	sp.mu.Unlock()

	go func() {
		defer sp.shutdownWg.Done()
		for {
			select {
			case msg := <-sub.ch:
				cb(msg)
			case <-sub.stop:
				return
			case <-sp.closeCh:
				return
			}
		}
	}()

	return &subscription{
		sp:         sp,
		subject:    subject,
		subscriber: sub,
	}
}

```
## Метод publish

Записывает в канал подпсичика переданное сообщение

```golang
func (sp *subPub) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	if sp.closed {
		return nil
	}
	for _, sub := range sp.subscribers[subject] {
		select {
		case sub.ch <- msg:
		default:

		}
	}
	return nil
}
```

## Метод close
Закрывает канал closeCh сигнализирующий, что сервис больше не принимает новых подписчиков и сообщений. Затем поочередно завершает работу всех подписчиков. Выходи из функции при завершении всех подпсчиков, либо при завершении контекста.
```golang
func (sp *subPub) Close(ctx context.Context) error {
	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.closed = true
	close(sp.closeCh)
	for _, subs := range sp.subscribers {
		for _, sub := range subs {
			close(sub.stop)
		}
	}
	sp.mu.Unlock()

	done := make(chan struct{})
	go func() {
		sp.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

Реализация библиотеки  [subpub.go](./internal/pkg/subpub/subpub.go)

2. Для удобства имплементации в GRPC server реализован [SubPuber](./internal/domain/service/service.go), использующий доменные модели. Например если захотим сменить библиотеку subpub на другую, у нас не будет зависимостей от subpub в GRPC хендлерах. Или если нам потребует дополнительные действия с сообщением, например сохранять в бд или как-то его изменять. 

3. Реализация GRPC handler [handlers.go](./internal/app/grpchandler/handlers.go)


# Запуск сервиса

1. В проекте есть конфигурационный файл [local.yaml](./config/local.yaml) , в нем укажите host и port котрые должен слушать сервер.

2. Для запуска проекта есть Makefile с командой run. Чтобы запустить воспользуйтесь:
```bash 
    make run
```

# Примеры отправки grpc запросов с использованием grpcurl

В Makefile прописана команда для установки grpcurl

Для отправки и получения сообщений воспользуйтес командами
```bash
grpcurl -d '{"key": "test", "data": "hello"}' -plaintext localhost:8083 PubSub/Publish
```

```bash
grpcurl -d '{"key": "test"}' -plaintext  localhost:8083 PubSub/Subscribe
```

# Дополнительно 

1. В проекте есть logger [logger.go](./internal/pkg/logger/logger.go)
2. В проекте используется dependency injection при инициализации grpc сервера мы можем подменять реализацию subpub если она соответствует доменным моделям.
3. В функции main файла main.go реализован graceful shutdown при поступлении сигналов OS. 
    