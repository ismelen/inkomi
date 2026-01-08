package PercentShower

import (
	"fmt"
	"strings"
	"sync"
)

const BAR_WIDTH = 40

type logger struct {
	values map[string]progressionValue
	Chan   chan string
	sync.RWMutex
}

type progressionValue struct {
	total, current int
}

func New(cap int) *logger {
	return &logger{
		values: make(map[string]progressionValue),
		Chan:   make(chan string, cap*2),
	}
}

func (this *logger) AddField(label string, total int) {
	this.Lock()
	this.values[label] = progressionValue{
		current: 0,
		total:   total,
	}
	this.Unlock()
}

func (this *logger) RemoveField(label string) {
	this.Lock()
	delete(this.values, label)
	this.Unlock()
}

func (this *logger) RunAsync(firstMsg string) {
	go func() {
		for msg := range this.Chan {
			this.RLock()
			this.print(firstMsg)
			for label, value := range this.values {
				if label == firstMsg {
					continue
				}
				if label == msg {
					value.current++
					this.values[label] = value
				}
				this.print(label)
			}
			this.moveCursorUp(len(this.values))
			this.RUnlock()
		}
	}()
}

func (this *logger) print(msg string) {
	values := this.values[msg]

	percent := float64(values.current) / float64(values.total)
	completed := int(percent * BAR_WIDTH)

	bar := strings.Repeat("█", completed) +
		strings.Repeat("░", BAR_WIDTH-completed)

	fmt.Printf("(%s): [%s] %d%%\033[K\n", msg, bar, int(percent*100))
}

func (this *logger) moveCursorUp(levels int) {
	fmt.Printf("\033[%dF", levels)
}
