# Hailstorm
CLI tool for load testing with yaml config files.
According to the script from the configuration file, creates the specified number of virtual users per minute. Users Reproduce the sequence of actions specified in the **"steps"** array. It is also possible to substitute variables in **"path"**, **"body"** and **"headers"** in place of special labels, such as **{{ variable }}**. Values of static variables from the set **"variables"**, values-increments from the set **"increments"** or values **captured** from the response to the previous step can be used for substitution.

# Build
```
git clone https://github.com/sokolovskiyma/Hailstorm.git
go build *.go
```

# Usage
To run with a configuration file in an arbitrary directory
```
./hailstorm /pth/to/config.yml
```
To run with config.yml in same directory
```
./hailstorm 
```

# Simple config
```
#First phase
- name: 'Test-1'               #Name of the phase
  url: 'http://127.0.0.1:8888' #Base url
  time: 10                     #Time for stage
  timeout: 10                  #Custom timeout for response in seconds (default 120s)
  steps:                       #Array of virtual user steps
      #The first step of the virtual user
      - method: 'POST'         #Request method
        path: '/test'          #Request path
        #The body of the request with variables for substitution (e.g. {{ catchValue }}, {{ incTwo }})
        body: '{"firstValue": {{ incTwo }}, "secondValue": {"thirdValue": ["test", 123, "{{ catchValue }}"], "thirdAndHalfValue": "thirdAndHalf"}, "lastValue": 666}'
        headers:               #Set of request heasers
            X-Custom-Header: 'AnotherValue'
            X-Another-Custom-Header: '{{ valueForHeader }}'
        catch:                 #Variables to capture from server response.
                               #In the future, they can be used for substitution (e.g. {{ catchValue }}, {{ valueForHeader }})
            valueForHeader: 'secondValue.thirdAndHalfValue'
            catchValue: 'secondValue.thirdValue.0'
      #The second step user
      - method: 'GET'
        path: '/{{ catchValue }}'
        headers:
            X-Custom-Header: 'myvalue'
  variables:                   #Variables for substitution, the values in them can be updated using the block "catch"
      catchValue: 'test'
  datasets:                    #Datasets - text files with comma - separated values
                               #You can also use for substitution ({{ test_list1 }})
      test_list1: 
          file: './datasets/numbers.csv'
          mode: 'sequence'     #If you select the "sequence" mode, the values for the substitution will be used sequentially, starting from the first
      test_list2:
          file: './datasets/words.csv'
          mode: 'random'       #if you select the "random" mode, the values for the substitution will be used in random order
  increments:                  #Increments for substitution ({{ incTwo }})
      incOne:
          start: 0             #Initial value
          step: 1              #The value by which the increment increases with each use
      incTwo:
          start: 15
          step: 5
  load:                        #The parameters of the load (virtual users per second)
      from: 10                 #Initial value
      to: 50                   #Load target
      ramp: 10                 #The value by which the number of virtual users will increase
#Second phase
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
```
 
 # TODO
- [ ] Binary release
- [ ] README
- [ ] Datasets

