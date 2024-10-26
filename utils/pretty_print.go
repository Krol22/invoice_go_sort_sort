package utils

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(data any) string {
    jsonData, err := json.MarshalIndent(data, "", "    ")
    if err != nil {
        fmt.Printf("Error marshaling: %v\n", err)
        return ""
    }
    return string(jsonData)
}
