package main

import "time"

type Config []Phase

type Phase struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	Time       int    `yaml:"time"`
	Timeout    int    `yaml:"timeout"`
	Steps      []Step
	Variables  map[string]string `yaml:"variables"`
	Increments map[string][]int  `yaml:"increments"`
	Datasets   map[string]struct {
		File    string `yaml:"file"`
		Mode    string `yaml:"mode"`
		Data    []string
		Positon string
	} `yaml:"datasets"`
	Load struct {
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
