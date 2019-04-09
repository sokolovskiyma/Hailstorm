package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	yaml "gopkg.in/yaml.v2"
)

// Читает конфигурационный файл
func readConfig() error {
	var confile []byte
	var err error

	// Берет конфигурационный файл либо из первого аргуиента командной строки, либо стандартный "./config.yml"
	if len(os.Args) > 1 {
		confile, err = ioutil.ReadFile(os.Args[1])
	} else {
		confile, err = ioutil.ReadFile("./config.yml")
	}
	if err != nil {
		return err
	}

	// Считывает файл в структуру типа Config
	err = yaml.Unmarshal(confile, &config)
	if err != nil {
		return err
	}

	// Проверяет валидность конфигурации, проставляет незаполненные значения
	err = validateConfig()
	if err != nil {
		return err
	}

	// Читает файлы с наборами данных, разбиват их по разделителю и записывет в массивы
	// for {
	// TODO
	// }

	return err
}

func validateConfig() error {
	var err error
	for index, phase := range config {
		if phase.Load.Ramp == 0 {
			config[index].Load.Ramp = 1
		}
		if phase.Load.To == 0 {
			config[index].Load.To = phase.Load.From
		}
		for key, _ := range phase.Increments {
			if len(phase.Increments[key]) < 2 {
				phase.Increments[key] = append(phase.Increments[key], 1)
			}
		}
	}
	return err
}

func replaceVar(stringToReplace string, variables map[string]string, increments map[string]int) string {
	re := regexp.MustCompile(`{{ \w+ }}`)
	for _, match := range re.FindAllString(stringToReplace, -1) {
		if value, ok := variables[match[3:len(match)-3]]; ok {
			stringToReplace = strings.Replace(stringToReplace, match, value, -1)
		}
		if value, ok := increments[match[3:len(match)-3]]; ok {
			stringToReplace = strings.Replace(stringToReplace, match, strconv.Itoa(value), -1)
		}
	}
	return stringToReplace
}

// Вытягиеват из json ответа одно значение
func catchValues(catch map[string]string, json string, variables *map[string]string) error {
	var err error
	// Валидирует json
	if !gjson.Valid(json) {
		err = errors.New("invalid json")
		return err
	}
	for key, value := range catch {
		(*variables)[key] = gjson.Get(json, value).String()
	}
	return err
}

// Выводит результаты фазы
func printResult(result Result) {
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
