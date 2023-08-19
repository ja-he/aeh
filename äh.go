package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {

	flag.Parse()
	prompt := flag.Arg(0)
	strings.ReplaceAll(prompt, `"`, `\"`)

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
	body := strings.NewReader(fmt.Sprintf(
		`{"model":"gpt-4","messages":[{"role":"user","content":"%s"}],"temperature":0.7}`,
		prompt,
	))

	// create a new request with JSON body data
	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		body,
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


	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading HTTP response body (%s)\n", err.Error())
		os.Exit(1)
	}

	responseBodyMap := make(map[string]any)
	json.Unmarshal(responseBodyBytes, &responseBodyMap)
	gptAnswer := responseBodyMap["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
	fmt.Println(gptAnswer)

}
