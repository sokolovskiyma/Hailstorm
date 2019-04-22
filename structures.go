package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type Config []Phase

type Phase struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	Time       int    `yaml:"time"`
	Timeout    int    `yaml:"timeout"`
	Steps      []Step
	Variables  map[string]string     `yaml:"variables"`
	Increments map[string]*Incriment `yaml:"increments"`
	Datasets   map[string]Dataset    `yaml:"datasets"`
	Load       struct {
		From int `yaml:"from"`
		To   int `yaml:"to"`
		Ramp int `yaml:"ramp"`
	} `yaml:"load"`
}

type Step struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
	Catch   map[string]string `yaml:"catch"`
}

type Dataset struct {
	File    string `yaml:"file"`
	Mode    string `yaml:"mode"`
	Data    []string
	Positon int
}

type Incriment struct {
	Start int `yaml:"start"`
	Step  int `yaml:"step"`
}

type Result struct {
	ScenariesCount int
	RequestCount   int
	Duration       time.Duration
	Statuses       map[string]int
	Latencys       []time.Duration
	AverageRPS     float32
	Latency        struct {
		Min     time.Duration
		Max     time.Duration
		Average time.Duration
	}
}

func (step *Step) replaceVar(phase int, variables map[string]string) {
	re := regexp.MustCompile(`{{ \w+ }}`)
	mu.Lock()
	for key, _ := range step.Headers {

		// fatal error: concurrent map writes
		// fatal error: concurrent map read and map write
		step.Headers[key] = replaceVar(re, step.Headers[key], phase, variables)
	}
	step.Body = replaceVar(re, step.Body, phase, variables)
	step.Path = replaceVar(re, step.Path, phase, variables)
	mu.Unlock()
}

func (dataset *Dataset) Pharse() error {
	// var err error
	byteData, err := ioutil.ReadFile(dataset.File)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	stringData := string(byteData)
	for _, element := range strings.Split(stringData, ";") {
		dataset.Data = append(dataset.Data, strings.TrimSpace(element))
	}
	// fmt.Println(dataset.Data)
	return err
}

func (dataset *Dataset) GetValue() string {
	var answer string
	if dataset.Mode == "sequence" {
		mu.Lock()
		answer = dataset.Data[dataset.Positon]
		dataset.Positon++
		mu.Unlock()
	} else if dataset.Mode == "random" {
		answer = dataset.Data[rand.Intn(len(dataset.Data))]
	}
	return answer
	// TODO
}

func (inc *Incriment) GetAntTick() int {
	mu.Lock()
	defer mu.Unlock()
	tempValue := inc.Start
	inc.Start += inc.Step
	return tempValue
	// TODO

}

//////////////////////////////////////////////////

func (result *Result) increaseScenaries() {
	mu.Lock()
	defer mu.Unlock()
	result.ScenariesCount++
}

func (result *Result) increaseRequests() {
	mu.Lock()
	defer mu.Unlock()
	result.RequestCount++
}

func (result *Result) increaseStatuses(status string) {
	mu.Lock()
	defer mu.Unlock()
	result.Statuses[status]++
}

func (result *Result) appendLatency(latency time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	result.Latencys = append(result.Latencys, latency)
}

// Выводит результаты фазы
func (result *Result) Print() {
	// Инициализирует минимальное время первым значением, чтобы было с чем сравнивать
	// result.Latency.Min = 0
	for _, latency := range result.Latencys {
		result.Latency.Average += latency
		if latency > result.Latency.Max {
			result.Latency.Max = latency
		}
		if latency < result.Latency.Min || result.Latency.Min == 0 {
			result.Latency.Min = latency
		}
	}
	fmt.Print("\r\033[K")
	fmt.Printf("Scenaries complited: %d\n", result.ScenariesCount)
	fmt.Printf("Request sent:        %d\n", result.RequestCount)
	fmt.Printf("Elapsed time:        %v\n", result.Duration)
	fmt.Printf("AverageRPS:          %.2f\n", result.AverageRPS)
	fmt.Printf("Codes:\n")
	for code, count := range result.Statuses {
		fmt.Printf("    %s:             %d\n", code, count)
	}
	if len(result.Latencys) == 0 {
		fmt.Printf("There is no latency info\n")
	} else {
		fmt.Printf("Request latensys:\n")
		fmt.Printf("    Min:             %v\n", result.Latency.Min)
		fmt.Printf("    Max:             %v\n", result.Latency.Max)
		fmt.Printf("    Average:         %v\n", result.Latency.Average/time.Duration(len(result.Latencys)))
	}
}

//////////////////////////////////////////////////
