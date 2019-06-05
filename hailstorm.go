package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var config Config
var wg sync.WaitGroup
var client *http.Client
var mu sync.Mutex

func main() {
	err := readConfig()
	if err != nil {
		fmt.Printf("Config file unmarshal error - %s", err)
		return
	}
	for index, phase := range config {
		if phase.Timeout != 0 {
			client = &http.Client{Timeout: time.Duration(phase.Timeout) * time.Second}
		} else {
			client = &http.Client{Timeout: 120 * time.Second}
		}

		phaseTitle := "│    Phase - " + config[index].Name + "    │"
		fmt.Print("\n┌", strings.Repeat("─", utf8.RuneCountInString(phaseTitle)-2), "┐\n")
		fmt.Printf(phaseTitle)
		fmt.Print("\n└", strings.Repeat("─", utf8.RuneCountInString(phaseTitle)-2), "┘\n")
		fmt.Print("\n")
		reproducePhase(index)
	}

}

func reproducePhase(phase int) {
	var result Result
	result.Statuses = map[string]int{}

	degrees := (config[phase].Load.To-config[phase].Load.From)/config[phase].Load.Ramp + 1
	timePerDegree := time.Duration(float64(config[phase].Time) / float64(degrees) * float64(time.Second))
	curRPS := config[phase].Load.From

	now := time.Now()
	for i := 0; i < degrees; i++ {
		// fmt.Printf("curRPS  - %v | time elapsed - %v\n", curRPS, time.Since(now).Round(time.Second))
		interval := time.Duration(1000/curRPS) * time.Millisecond
		// fmt.Printf("%v\n", interval)
		ticker := time.NewTicker(interval)
		subTicker := time.NewTicker(250 * time.Millisecond)
		go func() {
			for range ticker.C {
				wg.Add(1)
				// Запускает виртуальных юзеров, каждого в своей горутине
				go func(phase int, result *Result) {
					preproduceSteps(phase, result)
				}(phase, &result)
			}
		}()
		// go func() {
		// 	for range subTicker.C {
		// 		fmt.Printf("\rVirtual users per second - %v | Scenaries run: %d | Request send: %d | Time elapsed: %v", curRPS, result.ScenariesCount, result.RequestCount, time.Since(now).Round(time.Second))
		// 	}
		// }()
		time.Sleep(timePerDegree)

		ticker.Stop()
		subTicker.Stop()
		curRPS += config[phase].Load.Ramp
	}
	// fmt.Printf("Elapsed time - %v\n", time.Since(now))
	wg.Wait()
	result.Duration = time.Since(now).Round(time.Second)
	result.AverageRPS = float32(result.ScenariesCount) / float32(time.Since(now)/time.Second)
	result.Print()
}

func preproduceSteps(phase int, result *Result) {
	//
	localSteps := config[phase].Steps
	localVariables := config[phase].Variables
	defer wg.Done()
	result.increaseScenaries()
	for step, _ := range localSteps {
		localSteps[step].replaceVar(phase, localVariables)
		doReq(&localSteps[step], phase, &localVariables, result)
	}
}

// Делает запрос с текущими настройками
func doReq(step *Step, phase int, variables *map[string]string, result *Result) {
	var err error
	defer func() {
		if err != nil {
			result.Statuses["Errors"]++
			//panic(err)
		}
	}()

	result.increaseRequests()

	req, err := http.NewRequest(step.Method, config[phase].URL+step.Path, bytes.NewBuffer([]byte(step.Body)))
	if err != nil {
		fmt.Println(err.Error())
	}
	for key, value := range step.Headers {
		req.Header.Set(key, value)
	}

	now := time.Now()

	// fmt.Println("---", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		result.increaseStatuses("Errors")
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	result.appendLatency(time.Since(now))
	result.increaseStatuses(strconv.Itoa(resp.StatusCode))

	if resp.StatusCode == 404 {
		fmt.Println()
	}

	// Если статус ответа == 200
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		// Если тело ответа не пустое
		if stringBody := string(body); stringBody != "" {
			err = catchValues(step.Catch, stringBody, variables)
			if err != nil {
				return
			}
		}

	}

}
