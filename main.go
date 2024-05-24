package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	language = "en"
	country  = "us"
)

var (
	alphanumberic = append(rangeChar('a', 'z'), rangeChar('0', '9')...)
	characters    = rangeChar('a', 'z')
	headers       = map[string]string{
		"accept":          "*/*",
		"accept-language": "en-US,en;q=0.9,de-DE;q=0.8,de;q=0.7,en-DE;q=0.6",
		"content-type":    "application/x-www-form-urlencoded;charset=UTF-8",
	}
	params = map[string]string{
		"rpcids":      "teXCtc",
		"source-path": "/store/games",
		"f.sid":       "6591558633047063482",
		"bl":          "boq_playuiserver_20240521.08_p0",
		"hl":          language,
		"gl":          country,
		"authuser":    "0",
		"soc-app":     "121",
		"soc-platform": "1",
		"soc-device":  "1",
		"_reqid":      "978784",
		"rt":          "c",
	}
)

func main() {
	combinations := combinationGenerator(characters, 2)
	fmt.Println("len of char list: ", len(combinations))

	var resultList []map[string]string
	counter := 0

	for _, term := range combinations {
		counter++
		fmt.Printf("%d/%d\n", counter, len(combinations))
		response := callGPSuggestAPI(term)
		var responseBody []byte
		responseBody, _ = io.ReadAll(response.Body)
		response.Body.Close()
		var nestedKeywords [][]interface{}
		err := json.Unmarshal([]byte(strings.Split(string(responseBody), "\n")[3]), &nestedKeywords)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			continue
		}
		keywords := nestedKeywords[0][2].([]interface{})[0].([]interface{})
		for _, k := range keywords {
			keyword := k.([]interface{})[0].(string)
			resultList = append(resultList, map[string]string{
				"search_term":       term,
				"suggested_keyword": keyword,
			})
		}
	}

	fmt.Println("Len result list:", len(resultList))
}

func rangeChar(start, end rune) []string {
	var chars []string
	for c := start; c <= end; c++ {
		chars = append(chars, string(c))
	}
	return chars
}

func combinationGenerator(characters []string, length int) []string {
	if length == 1 {
		return characters
	}
	var combinations []string
	for _, char := range characters {
		subCombinations := combinationGenerator(characters, length-1)
		for _, subCombination := range subCombinations {
			combinations = append(combinations, char+subCombination)
		}
	}
	return combinations
}

func callGPSuggestAPI(searchTerm string) *http.Response {
	data := fmt.Sprintf(`f.req=%5B%5B%5B%22teXCtc%22%2C%22%5Bnull%2C%5B%5C%22%s%5C%22%5D%2C%5B10%5D%2C%5B2%2C1%5D%2C4%5D%22%2Cnull%2C%22generic%22%5D%5D%5D&at=AHEfX0v2tcQRorogvL7a9C2LqJR3%3A1716493977115&`, searchTerm)
	req, err := http.NewRequest("POST", "https://play.google.com/_/PlayStoreUi/data/batchexecute", bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	return resp
}
