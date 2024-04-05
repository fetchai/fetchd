package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func findPaths(data interface{}, target string, currentPath string, paths *map[string]string) {
	switch val := data.(type) {
	case map[string]interface{}:
		for key, value := range val {
			newPath := fmt.Sprintf("%s.%s", currentPath, key)
			if strings.Contains(key, target) {
				(*paths)[newPath] = key
			}
			findPaths(value, target, newPath, paths)
		}
	case []interface{}:
		for i, value := range val {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			valueStr := fmt.Sprintf("%v", value)
			keyStr := fmt.Sprintf("%v", i)
			if strings.Contains(keyStr, target) {
				(*paths)[newPath] = keyStr
				findPaths(value, target, newPath, paths)
				break
			}
			if strings.Contains(valueStr, target) {
				(*paths)[newPath] = valueStr
				findPaths(value, target, newPath, paths)
				break
			}
		}
	case string:
		if strings.Contains(val, target) {
			(*paths)[currentPath] = val
		}
	}
}

func printPaths(data interface{}, target string) {
	var paths = make(map[string]string)
	findPaths(data, target, "", &paths)
	fmt.Println("Paths containing", target)
	for path, value := range paths {
		if len(value) > 200 {
			value = value[:200] + "..."
		}
		if value == target {
			fmt.Println(path)
		} else {
			fmt.Println(path, "=>", value)
		}
	}
}

func main() {
	jsonData, err := ioutil.ReadFile("genesis.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	replacements := map[string]string{
		`("denom"\s*:\s*)"afet"`:           `${1}"aasi"`,
		`("[^"]*"\s*:\s*"[^"]*)(\d+)afet"`: `${1}aasi${2}`,

		// TODO:
		//"base" : "afet" -> "base" : "aasi",
		//"denom" : "XFET" -> "denom" : "XASI",
		//"display" : "FET" -> "display" : "ASI",
		//"name" : "FET" -> "name" : "ASI",
		//"symbol" : "FET" -> "symbol" : "ASI",
		// denom_metadata stuff,
		// bond_denom: "afet" -> "aasi",
		// mint_denom: "afet" -> "aasi",
		// denom_traces,

	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	newData := replaceValues(data, replacements, "")

	modifiedJSON, err := json.MarshalIndent(newData, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	if err := ioutil.WriteFile("modified_data.json", modifiedJSON, 0644); err != nil {
		fmt.Println("Error writing modified JSON to file:", err)
		return
	}
	fmt.Println("Modified JSON written to modified_data.json")
}

func isMatching(str string, replacements map[string]string) bool {
	for pattern, _ := range replacements {
		re := regexp.MustCompile(pattern)
		if re.MatchString(str) {
			return true
		}
	}
	return false
}

// TODO: this needs to make the comparison against the map value the whole way down, not just when it reaches a string at the bottom of the tree
func replaceValues(data map[string]interface{}, replacements map[string]string, path string) map[string]interface{} {
	for key, value := range data {
		currentPath := path + "->" + key
		switch v := value.(type) {
		case string:
			data[key] = replaceString(v, replacements, currentPath)
		case map[string]interface{}:
			jsonString, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				return data
			}

			data[key] = replaceString(string(jsonString), replacements, currentPath)
			data[key] = replaceValues(v, replacements, currentPath)
		case []interface{}:
			for i, item := range v {
				switch itemValue := item.(type) {
				case string:
					data[key].([]interface{})[i] = replaceString(itemValue, replacements, currentPath)

				case map[string]interface{}:
					data[key].([]interface{})[i] = replaceValues(itemValue, replacements, currentPath)
				}
			}
		}
	}
	return data
}

func replaceString(s string, replacements map[string]string, path string) string {
	for pattern, replace := range replacements {
		re := regexp.MustCompile(pattern)
		if re.MatchString(s) {
			fmt.Printf("\n\n")
			fmt.Printf("Path: %s\n", path)
			var toPrint string
			if len(s) > 200 {
				toPrint = s[:200] + "..."
			} else {
				toPrint = s
			}
			fmt.Printf("Before:\n%v\n\n", toPrint)
			s = re.ReplaceAllString(s, replace)
			if len(s) > 200 {
				toPrint = s[:200] + "..."
			} else {
				toPrint = s
			}
			fmt.Println("After:\n", toPrint)
		}
	}
	return s
}
