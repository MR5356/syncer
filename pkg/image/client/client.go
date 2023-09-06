package client

import (
	"github.com/MR5356/syncer/pkg/image/config"
	"github.com/MR5356/syncer/pkg/image/task"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Client struct {
	taskList *task.List

	failedTaskList  *task.List
	succeedTaskList *task.List

	config *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		taskList: task.NewTaskList(),

		failedTaskList:  task.NewTaskList(),
		succeedTaskList: task.NewTaskList(),

		config: cfg,
	}
}

func (c *Client) Run() error {
	start := time.Now()

	var ch = make(chan struct{}, c.config.Proc)
	var wg = sync.WaitGroup{}

	taskList, err := task.GenerateSyncTaskList(c.config, ch)
	if err != nil {
		logrus.Fatalf("error generate sync task list: %s", err)
	}
	c.taskList = taskList

	logrus.Infof("run sync task with %d processes", c.config.Proc)

	for t := range c.taskList.Iterator() {
		//ch <- struct{}{}
		wg.Add(1)

		t := t
		go func() {
			logrus.Infof("start sync task: %s", t.Name())

			if err := retry.Do(
				t.Run,
				retry.Attempts(uint(c.config.Retries)),
				retry.Delay(0),
				retry.LastErrorOnly(true),
				retry.DelayType(retry.DefaultDelayType),
				retry.OnRetry(func(n uint, err error) {
					logrus.Warnf("%d/%d: retry %s with error %s", n+1, c.config.Retries, t.Name(), err)
				}),
			); err != nil {
				logrus.Errorf("run sync task %s failed: %s", t.Name(), err)
				c.failedTaskList.Add(t)
			} else {
				logrus.Infof("run sync task %s succeed", t.Name())
				c.succeedTaskList.Add(t)
			}

			//<-ch
			wg.Done()
		}()
	}
	wg.Wait()

	cost := time.Since(start).String()

	if c.failedTaskList.Length() > 0 {
		for t := range c.failedTaskList.Iterator() {
			logrus.Warnf("task %s failed", t.Name())
		}
	}
	logrus.Infof("image sync finished, %d/%d task failed, cost %s", c.failedTaskList.Length(), c.taskList.Length(), cost)
	return nil
}
