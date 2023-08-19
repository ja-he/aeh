package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
)

func main() {

	model := flag.String("m", "gpt-4", "the model to use")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: äh [flags] <prompt>\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	prompt := flag.Arg(0)
	if prompt == "" {
		fmt.Fprintf(os.Stderr, "no prompt given\n")
		flag.Usage()
		os.Exit(1)
	}

	// we will make a POST request corresponding to this cURL command:
	//
	//   curl https://api.openai.com/v1/chat/completions \
	//     -H "Authorization: Bearer ${OPENAI_API_KEY}" \
	//     -H 'Content-Type: application/json' \
	//     -d '{
	//           model":"gpt-4",
	//           messages":[
	//             {
	//               "role":"user",
	//               "content":"${PROMPT}"
	//             }
	//           ],
	//           "temperature":0.7
	//         }'

	// create a new request body
	requestBodyData := map[string]any{
		"model": *model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": prompt,
		}},
		"temperature": 0.7,
	}
	requestBodyBytes, _ := json.Marshal(requestBodyData)
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	// create a new request with JSON body data
	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		requestBodyReader,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating HTTP request (%s)\n", err.Error())
		os.Exit(1)
	}
	// set the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	// send via default http client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error doing HTTP request (%s)\n", err.Error())
		os.Exit(1)
	}

	// HTTP/2.0 200 OK
	// Access-Control-Allow-Origin: *
	// Alt-Svc: h3=":443"; ma=86400
	// Cache-Control: no-cache, must-revalidate
	// Cf-Cache-Status: DYNAMIC
	// Cf-Ray: <HEX>-FRA
	// Content-Type: application/json
	// Date: Sat, 19 Aug 2023 18:05:29 GMT
	// Openai-Model: gpt-4-0613
	// Openai-Organization: <ORG-STRING e.g. 'user-[...]'>
	// Openai-Processing-Ms: 2445
	// Openai-Version: 2020-10-01
	// Server: cloudflare
	// Strict-Transport-Security: max-age=15724800; includeSubDomains
	// X-Ratelimit-Limit-Requests: 200
	// X-Ratelimit-Limit-Tokens: 10000
	// X-Ratelimit-Remaining-Requests: 199
	// X-Ratelimit-Remaining-Tokens: 9973
	// X-Ratelimit-Reset-Requests: 300ms
	// X-Ratelimit-Reset-Tokens: 162ms
	// X-Request-Id: <16 Bytes as Hex>
	//
	//  {
	//    "id": "chatcmpl-<BASE64>",
	//    "object": "chat.completion",
	//    "created": 1692468326,
	//    "model": "gpt-4-0613",
	//    "choices": [
	//      {
	//        "index": 0,
	//        "message": {
	//          "role": "assistant",
	//          "content": "<RESPONSE TEXT>"
	//        },
	//        "finish_reason": "stop"
	//      }
	//    ],
	//    "usage": {
	//      "prompt_tokens": 15,
	//      "completion_tokens": 23,
	//      "total_tokens": 38
	//    }
	//  }

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "received non-200 HTTP-status-code (%d)\n", resp.StatusCode)
		resp.Write(os.Stderr)
		os.Exit(1)
	}

	remainingRequests := resp.Header.Get("X-Ratelimit-Remaining-Requests")
	remainingTokens := resp.Header.Get("X-Ratelimit-Remaining-Tokens")
	fmt.Fprintf(os.Stderr, "remaining requests: %s\n", remainingRequests)
	fmt.Fprintf(os.Stderr, "remaining tokens: %s\n", remainingTokens)

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading HTTP response body (%s)\n", err.Error())
		os.Exit(1)
	}

	responseBodyMap := make(map[string]any)
	json.Unmarshal(responseBodyBytes, &responseBodyMap)

	// try to get the answer, but recover from any panic this may cause
	var gptAnswer string
	var modelResponding string
	var totalTokensUsed int
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "response is not of expected shape (%v)\n", r)
				resp.Write(os.Stderr)
				os.Exit(1)
			}
		}()
		gptAnswer = responseBodyMap["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
		modelResponding = responseBodyMap["model"].(string)
		totalTokensUsed = int(math.Round(responseBodyMap["usage"].(map[string]any)["total_tokens"].(float64)))
	}()

	fmt.Println(gptAnswer)

	fmt.Fprintf(os.Stderr, "model: %s\n", modelResponding)
	fmt.Fprintf(os.Stderr, "total tokens used: %d\n", totalTokensUsed)

}
