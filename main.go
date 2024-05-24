package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	language = "en"
	country  = "us"
)

var (
	characters = rangeChar('a', 'z')
	headers    = map[string]string{
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
		if response == nil {
			continue
		}

		var responseBody []byte
		responseBody, _ = io.ReadAll(response.Body)
		response.Body.Close()
		var nestedResponse []interface{}
		err := json.Unmarshal([]byte(strings.Split(string(responseBody), "\n")[3]), &nestedResponse)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			continue
		}

		keywordsData, ok := nestedResponse[0].([]interface{})
		if !ok || len(keywordsData) < 3 {
			fmt.Println("Unexpected JSON structure")
			continue
		}

		keywordsList, ok := keywordsData[2].([]interface{})
		if !ok || len(keywordsList) == 0 {
			fmt.Println("Unexpected JSON structure")
			continue
		}

		keywords, ok := keywordsList[0].([]interface{})
		if !ok {
			fmt.Println("Unexpected JSON structure")
			continue
		}

		for _, k := range keywords {
			keywordData, ok := k.([]interface{})
			if !ok || len(keywordData) == 0 {
				continue
			}
			keyword, ok := keywordData[0].(string)
			if !ok {
				continue
			}
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
	data := url.Values{}
	data.Set("f.req", fmt.Sprintf(`[[["teXCtc","[null,[\"%s\"],[10],[2,1],4]",null,"generic"]]]`, searchTerm))
	data.Set("at", "AHEfX0v2tcQRorogvL7a9C2LqJR3:1716493977115")

	req, err := http.NewRequest("POST", "https://play.google.com/_/PlayStoreUi/data/batchexecute", bytes.NewBufferString(data.Encode()))
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
	req.Header.Add("Content-Length", fmt.Sprint(len(data.Encode())))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Received non-OK response: %s\n", resp.Status)
		return nil
	}
	return resp
}
