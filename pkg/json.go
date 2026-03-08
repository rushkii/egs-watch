package pkg

import "encoding/json"

func StringJSON(target interface{}) (string, error) {
	str, err := json.MarshalIndent(target, "", "  ")

	if (err) != nil {
		return "", err
	}

	return string(str), nil

}
