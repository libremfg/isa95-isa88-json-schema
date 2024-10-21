package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// This is a very specific script that took all the definitions within v2.0.0.dataType.schema.json and checked if it existed within the base file
// if it existed (which they all do) then it will copy the reference over, which saves a lot of tedious copy and pasting
func DTchangeReferences(newFile string, baseFile string) (int, error) {

	// Extract keys from the JSON file
	keys, err := DTextractKeys(newFile)
	if err != nil {
		return fmt.Println("Error extracting keys:", err)
	}

	// Get the $ref values from the base file
	refs, err := DTgetRefs(keys, baseFile)
	if err != nil {
		return fmt.Println("Error getting references:", err)
	}

	// Replace the $ref values in the output file
	if err := DTreplacesRefs(newFile, refs); err != nil {
		return fmt.Println("Error replacing references:", err)
	}

	return 0, nil
}

// Extracts the definitions keys from the dataType schema file
func DTextractKeys(jsonFile string) ([]string, error) {

	// read the json file
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v", err)
	}

	// unmarshall the data from reading the json file and map it to the struct in main
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}

	// Extract the $defs part of the JSON data
	defs, ok := jsonData["$defs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("$defs section not found or is not an object")
	}

	// create string array for mapping definitions from schema
	var keys []string
	for key := range defs {
		keys = append(keys, key)
	}

	return keys, nil
}

// Finds the references within the base schema file and creates a map containing them
func DTgetRefs(keys []string, baseFile string) (map[string]string, error) {

	// Read the JSON file
	data, err := os.ReadFile(baseFile)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v", err)
	}

	// Unmarshal JSON data
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}

	// Extract the $defs part of the JSON data
	defs, ok := jsonData["$defs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("$defs section not found or is not an object")
	}

	// Initialize the map to store keys and their corresponding $ref values
	refMap := make(map[string]string)
	for _, key := range keys {
		if def, found := defs[key]; found {
			// Check if $ref exists in the definition
			if defMap, ok := def.(map[string]interface{}); ok {
				if refValue, ok := defMap["$ref"].(string); ok {
					refMap[key] = refValue
				}
			}
		}
	}

	return refMap, nil
}

// This will write the references found in the base file to the dataType schema file
func DTreplacesRefs(outputFile string, refValues map[string]string) error {

	// Read the JSON file
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}

	// Unmarshal JSON data
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("error unmarshalling JSON data: %v", err)
	}

	// Check if $defs section exists
	defs, ok := jsonData["$defs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("$defs section not found or is not an object")
	}

	// Replace $ref values only within the $defs section
	updateRefs(defs, refValues)

	// Marshal the updated JSON data
	updatedData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling updated JSON data: %v", err)
	}

	// Write the updated JSON data back to the file
	if err := os.WriteFile(outputFile, updatedData, 0644); err != nil {
		return fmt.Errorf("error writing updated JSON file: %v", err)
	}

	return nil
}

// updateRefs recursively updates the $ref values in the $defs section
func updateRefs(data interface{}, refValues map[string]string) {

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if newRef, exists := refValues[key]; exists {
				// If the key exists in refValues, update the corresponding $ref
				if refMap, ok := value.(map[string]interface{}); ok {
					if _, refExists := refMap["$ref"]; refExists {
						fmt.Printf("Updating $ref for key: %s -> %s\n", key, newRef)
						refMap["$ref"] = newRef
					}
				}
			} else {
				// Recursively update nested maps and slices within the $defs section
				updateRefs(value, refValues)
			}
		}
	case []interface{}:
		for _, item := range v {
			updateRefs(item, refValues)
		}
	}
}
