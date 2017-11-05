package main

import (
	"encoding/json"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"math/rand"
	"strconv"
	"crypto/tls"
	"io"
	"log"
	"sort"
)

func main() {
	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		downloadFiles()
	}

	genStats()
}

func genStats() {
	files, err := ioutil.ReadDir("./tmp")
	if err != nil {
		log.Fatal(err)
	}

	stat := make(map[string]map[string]int)

	for _, file := range files {
		content, err := ioutil.ReadFile("./tmp/" + file.Name())

		if err == nil {
			var value interface{}
			err = json.Unmarshal(content, &value)

			if err == nil {
				castedValue := value.(map[string]interface{})
				apiVersion, ok := castedValue["api"].(string)

				if ok {
					if _, ok := stat[apiVersion]; !ok {
						stat[apiVersion] = make(map[string]int)
					}
					stat[apiVersion] = flatten(castedValue, stat[apiVersion], "")
				}
			}
		}
	}

	for key, value := range stat {
		keys := make([]string, len(value))
		idx := 0
		for k := range value {
			keys[idx] = k
			idx++
		}
		sort.Strings(keys)

		count := value["/api"]
		fmt.Println("# --- " + key + " --- " + "(percent: " + strconv.Itoa((count * 100) / len(files)) + "%, total: " + strconv.Itoa(count) + ")")
		fmt.Println("| key | count | percent |")
		fmt.Println("|:--|:--|--:|")
		for _, innerKey := range keys {
			percent := (value[innerKey] * 100) / count
			fmt.Println("| " + innerKey + " | " + strconv.Itoa(value[innerKey]) + " | " + strconv.Itoa(percent) + "% |")
		}
	}
}

func flatten(from map[string]interface{}, to map[string]int, prepend string) map[string]int {
	for key, value := range from {
		obj, isObject := value.(map[string]interface{})

		if isObject {
			to = flatten(obj, to, prepend + "/" + key)
		} else {
			to[prepend + "/" + key] = to[prepend + "/" + key] + 1
		}
	}

	return to
}

func isString(i interface{}) bool {
	_, ok := i.(string)

	return ok
}

func isFloat(i interface{}) bool {
	_, ok := i.(float64)

	return ok
}

func isBool(i interface{}) bool {
	_, ok := i.(bool)

	return ok
}

func isArray(i interface{}) bool {
	_, ok := i.([]interface{})

	return ok
}

func downloadFiles() {
	response, err := http.Get("https://spaceapi.fixme.ch/directory.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	var f interface{}
	err = json.Unmarshal(body, &f)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	os.Mkdir("tmp", 0755)
	for _, value := range f.(map[string]interface{}) {
		spaceResponse, err := client.Get(value.(string))

		if err == nil {
			out, err := os.Create("tmp/" + strconv.Itoa(rand.Intn(999999)) + ".json")
			if err == nil {
				io.Copy(out, spaceResponse.Body)
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
}
