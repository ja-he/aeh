// Package is a command-line utility for calling LLM APIs, currently only the
// OpenAI API.
// It is intended to be simple and straightforward first and foremost.
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
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var errorf func(msg string, args ...any)

func main() {

	stderrATTY := isatty.IsTerminal(os.Stderr.Fd())
	if stderrATTY && os.Getenv("AEH_ERR") != "false" {
		errorf = func(msg string, args ...any) { color.New(color.FgYellow).Fprintf(os.Stderr, msg, args...) }
	} else if os.Getenv("AEH_ERR") == "false" {
		errorf = func(msg string, args ...any) {}
	} else {
		errorf = func(msg string, args ...any) { fmt.Fprintf(os.Stderr, msg, args...) }
	}

	model := flag.String("m", "gpt-4", "the model to use")
	temperature := flag.Float64("t", 0.7, "the temperature to use (see <https://platform.openai.com/docs/api-reference/chat/create#chat/create-temperature>)")
	flag.Usage = func() {
		errorf("usage: äh [flags] <prompt>\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	prompt := flag.Arg(0)
	if prompt == "" {
		errorf("no prompt given\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(flag.Args()) > 1 {
		errorf("additional command line (non-flag) arguments specified, which is invalid")
		flag.Usage()
		os.Exit(1)
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		if stdinInput, err := io.ReadAll(os.Stdin); err != nil {
			errorf("could not read from STDIN (%s)\n", err.Error())
			os.Exit(1)
		} else if len(stdinInput) > 0 {
			// append the STDIN input to the prompt (if any) with a newline in between
			prompt += "\n\n"
			prompt += string(stdinInput)
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		errorf("OPENAI_API_KEY environment variable not set\n")
		os.Exit(1)
	}

	// create a simple spinner
	spinner := NewSpinner(
		[]string{`🌑`, `🌘`, `🌗`, `🌕`, `🌔`, `🌓`, `🌒`},
		100*time.Millisecond,
		func() func(string, ...any) {
			if stderrATTY && os.Getenv("AEH_SPIN") == "" || os.Getenv("AEH_SPIN") == "true" {
				return func(msg string, args ...any) { fmt.Fprintf(os.Stderr, msg, args...) }
			} else {
				return func(msg string, args ...any) {}
			}
		}(),
	)

	// catch SIGINT and SIGTERM
	// (SIGTERM is sent by `kill` by default and SIGINT by `Ctrl+C`)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			spinner.Stop()
			errorf("received signal %s\n", sig)
			os.Exit(1)
		}
	}()

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

	gptAnswer, err := queryGPT(*model, apiKey, prompt, *temperature, spinner)
	if err != nil {
		errorf("error querying:\n")
		errorf(err.Error()+"\n")
		os.Exit(1)
	}

	fmt.Println(gptAnswer)
}

// Spinner is a simple spinner that can be used to indicate that something is
// currently happening.
type Spinner struct {
	iterations         []string
	frequency          time.Duration
	printFn            func(string, ...any)
	internalTerminator chan string
}

// NewSpinner creates a new Spinner with the given iterations, frequency and
// print function.
func NewSpinner(
	iterations []string,
	frequency time.Duration,
	printFn func(string, ...any),
) *Spinner {
	return &Spinner{
		iterations:         iterations,
		frequency:          frequency,
		printFn:            printFn,
		internalTerminator: make(chan string),
	}
}

// Spin starts the spinner and returns a channel that will be closed when the
// spinner is stopped.
func (s *Spinner) Spin(terminator <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(s.internalTerminator)
		defer s.printFn("\r                      \r") // clear the line
		ticker := time.NewTicker(s.frequency)
		iterIndex := 0
		s.printFn("\033[?25l")
		defer s.printFn("\033[?25h")
		for {
			select {
			case <-terminator:
				close(done)
				return
			case <-s.internalTerminator:
				return
			case <-ticker.C:
				s.printFn("\r%s", s.iterations[iterIndex])
				iterIndex = (iterIndex + 1) % len(s.iterations)
			}
		}
	}()
	return done
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	s.internalTerminator <- "stop requested"
	<-s.internalTerminator
}

func queryGPT(model, token, prompt string, temperature float64, spinner *Spinner) (string, error) {
	// create a new request body
	requestBodyData := map[string]any{
		"model": model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": prompt,
		}},
		"temperature": temperature,
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
		return "", fmt.Errorf("error creating HTTP request (%s)", err.Error())
	}
	// set the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	waitDone := make(chan struct{})
	spinnerDone := spinner.Spin(waitDone)

	// send via default http client
	resp, err := http.DefaultClient.Do(req)
	close(waitDone)
	<-spinnerDone
	if err != nil {
		return "", fmt.Errorf("error doing HTTP request (%s)", err.Error())
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
		resp.Write(os.Stderr)
		return "", fmt.Errorf("received non-200 HTTP-status-code (%d)", resp.StatusCode)
	}

	remainingRequests := resp.Header.Get("X-Ratelimit-Remaining-Requests")
	remainingTokens := resp.Header.Get("X-Ratelimit-Remaining-Tokens")
	errorf("remaining requests: %s\n", remainingRequests)
	errorf("remaining tokens: %s\n", remainingTokens)

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errorf("error reading HTTP response body (%s)\n", err.Error())
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
				errorf("response is not of expected shape (%v)\n", r)
				resp.Write(os.Stderr)
				os.Exit(1)
			}
		}()
		gptAnswer = responseBodyMap["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
		modelResponding = responseBodyMap["model"].(string)
		totalTokensUsed = int(math.Round(responseBodyMap["usage"].(map[string]any)["total_tokens"].(float64)))
	}()


	errorf("model: %s\n", modelResponding)
	errorf("total tokens used: %d\n", totalTokensUsed)

	return gptAnswer, nil
}
