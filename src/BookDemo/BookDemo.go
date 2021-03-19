package BookDemo

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Runner struct {
	interrupt chan os.Signal
	complete  chan error
	timerout  <-chan time.Time
	tasks     []func(int)
}

var ErrTimeOut = errors.New("received timeout")

var ErrInterupt = errors.New("received interrupt")

func New(r *Runner) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timerout:  time.After(100),
	}
}

func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) Start() error {
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-r.timerout:
		return ErrTimeOut
	}
}

func (r *Runner) run() error {
	for id, task := range r.tasks {
		if r.goInterrupt() {
			return ErrInterupt
		}
		task(id)
	}
	return nil
}

func (r *Runner) goInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}

type Worker interface {
	Task()
}

type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

func NewPool(maxGoroutines int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}
	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				w.Task()
			}
			p.wg.Done()
		}()
	}
	return &p
}

func (p *Pool) Run(w Worker) {
	p.work <- w
}

func (p *Pool) Shutdown() {
	close(p.work)
	p.wg.Wait()
}

func DoPanic2() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("do recover err ", err)
		}
	}()
	fmt.Println("11111")
	m := make([]byte, 0)
	fmt.Println("do wonrg,", m[100])
	fmt.Println("22222")

}

func DoPanic() {
	fmt.Println("0000000")
	DoPanic2()
	fmt.Println("33333333")
}
