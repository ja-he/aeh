package main

import (
	"flag"
	"fmt"
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
		panic(err)
	}
	// set the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	// send via default http client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	resp.Write(os.Stdout)
}
