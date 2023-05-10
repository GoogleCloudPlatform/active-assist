// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.```


// This package was necessary because slack verification
// requires the string to be in item:value form.
// However golang converts to {"item":"value"}
// Using the JSON package should work for this however...
// The JSON package unmarshals in any random order. 
// And since we are doing a hash to ensure verification
// We need it in the same order as it arrived from slack

// I do not think this code covers all edge cases. 

package utils

import (
    "errors"
	"strconv"
    "fmt"
)

func ParseJSONToMap(jsonStr string) (map[string]interface{}, error) {
    var result map[string]interface{}

    // Check if the input string is empty
    if jsonStr == "" {
        return nil, errors.New("input string is empty")
    }

    // Convert the input string to a byte slice
    jsonBytes := []byte(jsonStr)

    // Define a stack to keep track of nested objects and arrays
    var stack []interface{}

    // Initialize the current map to the result map
    var currentMap map[string]interface{} = result

	LogPrint(1, "Body: %s", string(jsonBytes))
    // Loop through the byte slice and parse the JSON data
    for i := 0; i < len(jsonBytes); i++ {
        char := jsonBytes[i]
        switch char {
        case '{':
            // Push a new map onto the stack
            stack = append(stack, currentMap)

            // Create a new map and set it as the current map
            newMap := make(map[string]interface{})
            currentMap = newMap
        case '}':
            // Pop the previous map off the stack
            if len(stack) == 0 {
                return nil, errors.New("unexpected '}'")
            }
            lastMapIndex := len(stack) - 1
            lastMap := stack[lastMapIndex]
            stack = stack[:lastMapIndex]
		
			// If it's the last one append to stack.
			if i == len(jsonBytes)-1{
				stack = append(stack, currentMap)
			}else{
            // Set the current map to the previous map
            currentMap = lastMap.(map[string]interface{})
			}
        case '[':
            // Push a new slice onto the stack
            stack = append(stack, currentMap)

            // Create a new slice and set it as the current map value
            newSlice := make(map[string]interface{})
            currentMap = newSlice
        case ']':
            // Pop the previous slice off the stack
            if len(stack) == 0 {
                return nil, errors.New("unexpected ']'")
            }
            lastMapIndex := len(stack) - 1
            lastMap := stack[lastMapIndex]
            stack = stack[:lastMapIndex]

            // Set the current map value to the previous slice
            currentMap = lastMap.(map[string]interface{})
        case ':':
            // Skip the ':' character
        case ',':
            // Skip the ',' character
        case '"':
            // Parse a string value
            endIndex := i + 1
            for ; endIndex < len(jsonBytes); endIndex++ {
                if jsonBytes[endIndex] == '"' {
                    break
                }
                if jsonBytes[endIndex] == '\\' {
                    endIndex++
                }
            }
            if endIndex == len(jsonBytes) {
                return nil, errors.New("unexpected end of string")
            }
			valueStart := endIndex+1
			for ; valueStart < len(jsonBytes); valueStart++ {
				if jsonBytes[endIndex] == '"' {
					valueStart++
                    break
                }
			}
            key := string(jsonBytes[i+1 : endIndex])
            if len(key) > 0 {
                // Parse the value
                value, newIndex, err := parseJSONValue(jsonBytes[valueStart:])
                if err != nil {
					LogPrint(3, "Error Parsing Json Value: %v", err)
                    return nil, err
                }
				//LogPrint(1, "Key: %s, Value: %s", key, value)
				// I needs to start at the end of where the value ends + 1
                i = endIndex + newIndex + 1
                // Add the key-value pair to the current map
                currentMap[key] = value
            }
        default:
		}
		//LogPrint(1, "Current Map: %v", currentMap)		
	}
	// This isn't right yet... But I need a break. 
	// Stack shouldn't be an array, should be a dict from the start.
	// However this currently works and I'm tired of working on this. 
	value, ok := stack[0].(map[string]interface{})
	if !ok {
		LogPrint(3, "Issue with Stack conversion")
	}
	result = value
    return result, nil
}


func parseJSONValue(jsonBytes []byte) (interface{}, int, error) {
    var endIndex int
    var err error
    var value interface{}
	//LogPrint(1, "parseJSONValue jsonBytes[0]: %v", string(jsonBytes[0]))
    switch jsonBytes[0] {
    case '{':
        // Parse an object
        value, endIndex, err = parseJSONObject(jsonBytes)
    case '[':
        // Parse an array
        value, endIndex, err = parseJSONArray(jsonBytes)
    case '"':
        // Parse a string
        value, endIndex, err = parseJSONString(jsonBytes)
    default:
        // Parse a number or boolean value
        value, endIndex, err = parseJSONNumber(jsonBytes)
    }

    if err != nil {
        return nil, 0, err
    }
	//LogPrint(1, "Value: %v, EndIndex: %v", value, endIndex)
    return value, endIndex, nil
}

func parseJSONObject(jsonBytes []byte) (map[string]interface{}, int, error) {
    obj := make(map[string]interface{})
    i :=  1
    for ; i < len(jsonBytes); i++ {
        char := jsonBytes[i]
        if char == '}' {
            return obj, i + 1, nil
        } else if char == ',' {
            continue
        } else if char != '"' {
            return nil, 0, fmt.Errorf("expected '\"', found '%c'", char)
        }
        key, index, err := parseJSONString(jsonBytes)
        if err != nil {
            return nil, 0, err
        }
        i = index
        char = jsonBytes[i]
        if char != ':' {
            return nil, 0, fmt.Errorf("expected ':', found '%c'", char)
        }
        value, newIndex, err := parseJSONValue(jsonBytes[i+1:])
        if err != nil {
            return nil, 0, err
        }
        obj[key] = value
        i += newIndex + 1
    }
    return nil, 0, fmt.Errorf("expected '}', found end of input")
}


func parseJSONArray(jsonBytes []byte) ([]interface{}, int, error) {
    var result []interface{}

    // Define a stack to keep track of nested objects and arrays
    var stack []interface{}

    // Initialize the current slice to the result slice
    currentSlice := &result

    // Loop through the byte slice and parse the JSON data
    for i := 0; i < len(jsonBytes); i++ {
        char := jsonBytes[i]

        switch char {
        case '[':
            // Push a new slice onto the stack
            stack = append(stack, currentSlice)

            // Create a new slice and set it as the current slice
            newSlice := make([]interface{}, 0)
            *currentSlice = newSlice
        case ']':
            // Pop the previous slice off the stack
            if len(stack) == 0 {
                return nil, 0, errors.New("unexpected ']'")
            }
            lastSliceIndex := len(stack) - 1
            lastSlice := stack[lastSliceIndex]
            stack = stack[:lastSliceIndex]

            // Set the current slice to the previous slice
            currentSlice = lastSlice.(*[]interface{})
        case ':':
            // Skip the ':' character
        case ',':
            // Skip the ',' character
        case '"':
            // Parse a string value
            value, newIndex, err := parseJSONString(jsonBytes[i:])
            if err != nil {
                return nil, 0, err
            }
            i += newIndex
            // Add the value to the current slice
            *currentSlice = append(*currentSlice, value)
        case '{':
            // Parse an object value
            value, newIndex, err := parseJSONObject(jsonBytes[i:])
            if err != nil {
                return nil, 0, err
            }
            i += newIndex
            // Add the value to the current slice
            *currentSlice = append(*currentSlice, value)
        default:
            // Parse a number or boolean value
            value, newIndex, err := parseJSONValue(jsonBytes[i:])
            if err != nil {
                return nil, 0, err
            }
            i += newIndex
            // Add the value to the current slice
            *currentSlice = append(*currentSlice, value)
        }
    }

    return result, len(jsonBytes), nil
}


func parseJSONString(jsonBytes []byte) (string, int, error) {
    // Check if the input string is empty
    if len(jsonBytes) == 0 {
        return "", 0, errors.New("empty JSON string")
    }

    // Check if the input string is a valid string
    if jsonBytes[0] != '"' {
        return "", 0, errors.New("invalid JSON string")
    }

    // Parse the string value
    endIndex := 1
    for ; endIndex < len(jsonBytes); endIndex++ {
        if jsonBytes[endIndex] == '"' {
            break
        }
        if jsonBytes[endIndex] == '\\' {
            endIndex++
        }
    }
    if endIndex == len(jsonBytes) {
        return "", 0, errors.New("unexpected end of string")
    }
    str := string(jsonBytes[1:endIndex])
    return str, endIndex + 1, nil
}



func parseJSONNumber(jsonBytes []byte) (interface{}, int, error) {
    var result interface{}
    var i int
    numStr := ""

    // Loop through the byte slice to extract the number string
    for i = 0; i < len(jsonBytes); i++ {
        char := jsonBytes[i]
        if char == '.' || char == '-' || (char >= '0' && char <= '9') {
            numStr += string(char)
        } else {
            break
        }
    }

    // Try to parse the number as an integer
    intVal, err := strconv.Atoi(numStr)
    if err == nil {
        result = intVal
        return result, i, nil
    }

    // Try to parse the number as a float
    floatVal, err := strconv.ParseFloat(numStr, 64)
    if err == nil {
        result = floatVal
        return result, i, nil
    }

	LogPrint(3,"parseJSONNumber passed bytes: %v", string(jsonBytes))
    return nil, 0, fmt.Errorf("expected number, found '%s'", numStr)
}
