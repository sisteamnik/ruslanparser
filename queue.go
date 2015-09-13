package ruslanparser

import (
	"fmt"
	"sync"
	"time"
)

type Cmd struct {
	Query    string
	SearchBy string
	Offset   int
	Limit    int

	ResultNum int
	Result    []Book
	Error     error

	Done   chan struct{}  `json:"-"`
	isWait bool           `json:"-"`
	wg     sync.WaitGroup `json:"-"`
}

func NewCmd(wg sync.WaitGroup) *Cmd {
	return &Cmd{Done: make(chan struct{}), wg: wg}
}

func (c Cmd) String() string {
	return fmt.Sprintf("%s_%s_%d_%d", c.Query, c.SearchBy, c.Offset, c.Limit)
}

func (c *Cmd) WaitDone() {
	c.isWait = true
	<-c.Done
}

func (c *Cmd) IsWait() bool {
	return c.isWait
}

type Worker struct {
	Id int

	SsidMu    sync.Mutex
	Ssid      string
	SsidFetch time.Time

	queue *Queue
}

func NewWorker(id int, q *Queue) *Worker {
	return &Worker{Id: id, queue: q}
}

func (w *Worker) Run() {
	for cmd := range w.queue.cmdChan {
		w.Do(cmd)
		if cmd.IsWait() {
			cmd.Done <- struct{}{}
		}
		w.queue.wg.Done()
	}
}

func (w *Worker) Session() (string, error) {
	w.SsidMu.Lock()
	defer w.SsidMu.Unlock()
	if w.Ssid != "" && time.Since(w.SsidFetch) < 5*time.Minute {
		return w.Ssid, nil
	}
	return NewSessionSearchUrl(w.queue.BaseUrl + "?" + w.queue.IndexQuery)
}

func (w *Worker) Do(cmd *Cmd) {
	if c := w.queue.Cache; c != nil {
		e := c.Get(cmd)
		if e == nil {
			return
		} else {
			fmt.Println("NOT  HACHED")
		}
	}
	fmt.Println(w.Id, "here", "cmd", cmd.Query)
	ssid, e := w.Session()
	fmt.Println(w.Id, ssid)
	if e != nil {
		cmd.Error = e
		return
	}
	cmd.Result, cmd.ResultNum, cmd.Error = Search(cmd.Query, cmd.SearchBy, ssid, cmd.Offset, cmd.Limit)
	if c := w.queue.Cache; c != nil {
		e = c.Set(cmd)
		if e != nil {
			cmd.Error = e
			return
		}
	}
}

type QueueOpts struct {
	BaseUrl      string
	IndexQuery   string
	WorkerNumber int
	Cache        Cacher
}

type Queue struct {
	*QueueOpts

	cmdChan chan *Cmd
	wg      sync.WaitGroup
}

func NewQueue(opts *QueueOpts) *Queue {
	return &Queue{
		QueueOpts: opts,
		cmdChan:   make(chan *Cmd)}
}

func (ws *Queue) Run() {
	for i := 0; i < ws.WorkerNumber; i++ {
		w := NewWorker(i, ws)
		go w.Run()
	}
}

func (ws *Queue) Wait() {
	ws.wg.Wait()
}

func (ws *Queue) Do(cmd *Cmd) {
	ws.wg.Add(1)
	ws.cmdChan <- cmd
}

func (ws *Queue) Search(q string, offset, limit int, by ...string) ([]Book, int, error) {
	if limit == 0 {
		limit = 10
	}
	cmd := NewCmd(ws.wg)
	cmd.Query = q
	cmd.SearchBy = BY_TITLE
	cmd.Offset = offset
	cmd.Limit = limit
	if len(by) > 0 {
		cmd.SearchBy = by[0]
	}
	ws.Do(cmd)
	cmd.WaitDone()
	return cmd.Result, cmd.ResultNum, cmd.Error
}
