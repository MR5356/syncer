package task

import "sync"

type Task interface {
	Name() string
	Run() error
}

type List struct {
	lock sync.RWMutex
	list []Task
}

func NewTaskList() *List {
	return &List{
		lock: sync.RWMutex{},
		list: make([]Task, 0),
	}
}

func (l *List) Add(t Task) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.list = append(l.list, t)
}

func (l *List) Length() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	return len(l.list)
}

func (l *List) Iterator() <-chan Task {
	c := make(chan Task)
	go func() {
		l.lock.Lock()
		defer l.lock.Unlock()
		for _, t := range l.list {
			c <- t
		}
		close(c)
	}()
	return c
}
