package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getimg(input, filename string) (string, error) {
	url := config.MdToImgURL
	data := []byte(fmt.Sprintf(`{"md": "%s"}`, input))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Error reading response body: %v", err)
		}

		err = ioutil.WriteFile(filename, body, 0644)
		if err != nil {
			return "", fmt.Errorf("Error saving image: %v", err)
		}

		return filename, nil
	} else {
		return "", fmt.Errorf("Error response: %v", resp.Status)
	}
}
