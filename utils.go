package main

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	for phase, _ := range config {
		for _, dataset := range config[phase].Datasets {
			err = dataset.Pharse()
		}
	}

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
		// for key, _ := range phase.Increments {
		// 	if len(phase.Increments[key]) < 2 {
		// 		phase.Increments[key] = append(phase.Increments[key], 1)
		// 	}
		// }
	}
	return err
}

func replaceVar(re *regexp.Regexp, stringToReplace string, phase int, variables map[string]string) string {
	// re := regexp.MustCompile(`{{ \w+ }}`)
	for _, match := range re.FindAllString(stringToReplace, -1) {
		if value, ok := variables[match[3:len(match)-3]]; ok {
			stringToReplace = strings.Replace(stringToReplace, match, value, -1)
		}
		if value, ok := config[phase].Increments[match[3:len(match)-3]]; ok {
			stringToReplace = strings.Replace(stringToReplace, match, strconv.Itoa(value.Start), -1)
		}
		if value, ok := config[phase].Datasets[match[3:len(match)-3]]; ok {
			stringToReplace = strings.Replace(stringToReplace, match, value.GetValue(), -1)
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
		mu.Lock()
		(*variables)[key] = gjson.Get(json, value).String()
		mu.Unlock()
	}
	return err
}
