# Hailstorm
CLI tool for load testing with yaml config files

# Build
- go build *.go

# Usage

# Simple config
```- name: 'Test-1'
  url: 'http://127.0.0.1:8888'
  time: 10
  timeout: 10
  steps:
      - method: 'POST'
        path: '/test'
        body: '{"firstValue": {{ incTwo }}, "secondValue": {"thirdValue": ["test", 123, "{{ catchValue }}"], "thirdAndHalfValue": "thirdAndHalf"}, "lastValue": 666}'
        headers:
            X-Custom-Header: 'AnotherValue'
            X-Another-Custom-Header: '{{ valueForHeader }}'
        catch:
            valueForHeader: 'secondValue.thirdAndHalfValue'
            catchValue: 'secondValue.thirdValue.0'
      - method: 'GET'
        path: '/{{ catchValue }}'
        headers:
            X-Custom-Header: 'myvalue'
  variables:
      catchValue: 'test'
  datasets:
      test_list1: 
          file: './datasets/numbers.csv'
          mode: 'sequence'
      test_list2:
          file: './datasets/words.csv'
          mode: 'random'
  increments:
      incOne:
          start: 0
          step: 1
      incTwo:
          start: 15
          step: 5
  load:
      from: 10
      #to: 50
      #ramp: 10
- name: 'Test-2'
  url: 'http://127.0.0.1:8888'
  time: 60
  timeout: 10
  steps:
      - method: 'POST'
        path: '/test'
        body: '{"firstValue": {{ incTwo }}, "secondValue": {"thirdValue": ["test", 123, "{{ catchValue }}"], "thirdAndHalfValue": "thirdAndHalf"}, "lastValue": 666}'
        headers:
            X-Custom-Header: 'AnotherValue'
        catch:
            itVar: 'secondValue.thirdAndHalfValue'
            catchValue: 'secondValue.thirdValue.0'
      - method: 'GET'
        path: '/{{ catchValue }}'
        headers:
            X-Custom-Header: 'myvalue'
  variables:
      catchValue: 'test'
  increments:
      incOne:
          start: 0
          step: 1
      incTwo:
          start: 15
          step: 5
  load:
      from: 10
      to: 50
      #ramp: 10```
