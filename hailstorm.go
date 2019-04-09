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
var mu sync.Mutex
var client *http.Client

func main() {
	err := readConfig()
	if err != nil {
		panic(err)
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
	// fmt.Printf("degrees - %v\n", degrees)
	// fmt.Printf("curRPS  - %v\n", curRPS)
	// fmt.Printf("timePerDegree - %v\n", timePerDegree)
	// fmt.Println()

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
				go preproduceSteps(phase, &result)
			}
		}()
		go func() {
			for range subTicker.C {
				fmt.Printf("\rVirtual users per second - %v | Scenaries run: %d | Request send: %d | Time elapsed: %v", curRPS, result.ScenariesCount, result.RequestCount, time.Since(now).Round(time.Second))
				// float32(result.ScenariesCount)/float32(time.Since(now)/time.Second),
			}
		}()
		time.Sleep(timePerDegree)

		ticker.Stop()
		subTicker.Stop()
		curRPS += config[phase].Load.Ramp
	}
	// fmt.Printf("Elapsed time - %v\n", time.Since(now))
	wg.Wait()
	result.Duration = time.Since(now).Round(time.Second)
	result.AverageRPS = float32(result.ScenariesCount) / float32(time.Since(now)/time.Second)
	printResult(result)
}

/*
 *	TODO: Возможно стоит не предавать в функцию кучу аргументов по отдельности,
 *		  а просто номер фазы. И использовать номер как то так config[phaseNumber].Incriments
 */
func preproduceSteps(phase int, result *Result) {
	defer wg.Done()
	mu.Lock()
	result.ScenariesCount++
	mu.Unlock()
	localIncrements := make(map[string]int)
	mu.Lock()
	// Копирует инкременты локаально, чтобы их никто не "инкрементнул" раньше чем их значения будут использованны
	for key, value := range config[phase].Increments {
		localIncrements[key] = value[0]
	}
	// Увеличивает глобальное значение инкрементов на задоанное значение, после разблокирывает мьютекс
	for _, increment := range config[phase].Increments {
		increment[0] += increment[1]
	}
	mu.Unlock()
	for step, _ := range config[phase].Steps {
		for key, _ := range config[phase].Steps[step].Headers {
			config[phase].Steps[step].Headers[key] = replaceVar(config[phase].Steps[step].Headers[key], config[phase].Variables, localIncrements)
		}
		config[phase].Steps[step].Body = replaceVar(config[phase].Steps[step].Body, config[phase].Variables, localIncrements)
		config[phase].Steps[step].Path = replaceVar(config[phase].Steps[step].Path, config[phase].Variables, localIncrements)
		doReq(phase, step, result)
	}
	// mu.Lock()
	// result.ScenariesCount++
	// mu.Unlock()
}

// Делает запрос с текущими настройками
func doReq(phase, step int, result *Result) {
	var err error
	defer func() {
		if err != nil {
			result.Statuses["Errors"]++
			//panic(err)
		}
	}()
	mu.Lock()
	result.RequestCount++
	mu.Unlock()
	req, err := http.NewRequest(config[phase].Steps[step].Method, config[phase].URL+config[phase].Steps[step].Path, bytes.NewBuffer([]byte(config[phase].Steps[step].Body)))
	if err != nil {
		fmt.Println(err.Error())
	}
	for key, value := range config[phase].Steps[step].Headers {
		req.Header.Set(key, value)
	}

	now := time.Now()

	// fmt.Println("---", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		mu.Lock()
		result.Statuses["Errors"]++
		mu.Unlock()
		return
		//fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	mu.Lock()
	result.Latencys = append(result.Latencys, time.Since(now))
	result.Statuses[strconv.Itoa(resp.StatusCode)]++
	mu.Unlock()

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
			err = catchValues(config[phase].Steps[step].Catch, stringBody, &config[phase].Variables)
			if err != nil {
				return
			}
		}

	}

}
