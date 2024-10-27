package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-jsonnet/formatter"
	"github.com/spf13/cobra"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/message"
)

//go:embed playground.tmpl
var playgroundHTML string

func init() {
	rootCmd.AddCommand(playgroundCmd)
}

var playgroundCmd = &cobra.Command{
	Use:   "playground",
	Short: "start playground",
	Long:  `'substation playground' starts a local HTTP server for testing Substation configurations.`,
	RunE:  runPlayground,
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	statusCode := http.StatusOK

	var err interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		err = v["error"]
	case map[string]string:
		err = v["error"]
	}

	if err != nil {
		statusCode = http.StatusInternalServerError
		log.Printf("Error in request: %v", err)
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func runPlayground(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/test", handleTest)
	mux.HandleFunc("/demo", handleDemo)
	mux.HandleFunc("/fmt", handleFmt)
	mux.HandleFunc("/share", handleShare) // Add this line

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Substation playground is running on http://localhost:8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down playground...")
	return server.Shutdown(ctx)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		DefaultConfig string
		DefaultInput  string
		DefaultOutput string
		DefaultEnv    string
	}{
		DefaultConfig: "",
		DefaultInput:  "",
		DefaultOutput: "",
		DefaultEnv:    "",
	}

	// Check for shared data in query string
	sharedData := r.URL.Query().Get("share")
	if sharedData != "" {
		decodedData, err := base64.URLEncoding.DecodeString(sharedData)
		if err == nil {
			parts := strings.SplitN(string(decodedData), "{substation-separator}", 3)
			if len(parts) == 3 {
				data.DefaultConfig = parts[0]
				data.DefaultInput = parts[1]
				data.DefaultOutput = parts[2]
			}
		}
	}

	// If shared data is present, don't include environment variables
	if sharedData == "" {
		data.DefaultEnv = "# Add environment variables here, one per line\n# Example: KEY=VALUE"
	}

	tmpl := template.Must(template.New("index").Parse(playgroundHTML))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"config": confDemo,
		"input":  evtDemo,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, map[string]string{"error": "Method not allowed"})
		return
	}

	var request struct {
		Config string `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONResponse(w, map[string]string{"error": "Invalid request"})
		return
	}

	conf, err := compileStr(request.Config, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error compiling config: %v", err), http.StatusBadRequest)
		return
	}

	var cfg customConfig
	if err := json.Unmarshal([]byte(conf), &cfg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var output strings.Builder

	if len(cfg.Transforms) == 0 {
		output.WriteString("?\t[config error]\n")
		sendJSONResponse(w, map[string]string{"output": output.String()})
		return
	}

	if len(cfg.Tests) == 0 {
		output.WriteString("?\t[no tests]\n")
		sendJSONResponse(w, map[string]string{"output": output.String()})
		return
	}

	start := time.Now()
	failedTests := false

	for _, test := range cfg.Tests {
		cnd, err := condition.New(ctx, test.Condition)
		if err != nil {
			output.WriteString("?\t[test error]\n")
			sendJSONResponse(w, map[string]string{"output": output.String()})
			return
		}

		setup, err := substation.New(ctx, substation.Config{
			Transforms: test.Transforms,
		})
		if err != nil {
			output.WriteString("?\t[test error]\n")
			sendJSONResponse(w, map[string]string{"output": output.String()})
			return
		}

		tester, err := substation.New(ctx, cfg.Config)
		if err != nil {
			output.WriteString("?\t[config error]\n")
			sendJSONResponse(w, map[string]string{"output": output.String()})
			return
		}

		sMsgs, err := setup.Transform(ctx, message.New().AsControl())
		if err != nil {
			output.WriteString("?\t[test error]\n")
			sendJSONResponse(w, map[string]string{"output": output.String()})
			return
		}

		tMsgs, err := tester.Transform(ctx, sMsgs...)
		if err != nil {
			output.WriteString("?\t[config error]\n")
			sendJSONResponse(w, map[string]string{"output": output.String()})
			return
		}

		testPassed := true
		for _, msg := range tMsgs {
			if msg.HasFlag(message.IsControl) {
				continue
			}

			ok, err := cnd.Condition(ctx, msg)
			if err != nil {
				output.WriteString("?\t[test error]\n")
				sendJSONResponse(w, map[string]string{"output": output.String()})
				return
			}

			if !ok {
				output.WriteString(fmt.Sprintf("--- FAIL: %s\n", test.Name))
				output.WriteString(fmt.Sprintf("    message:\t%s\n", msg))
				output.WriteString(fmt.Sprintf("    condition:\t%s\n", cnd))
				testPassed = false
				failedTests = true
				break
			}
		}

		if testPassed {
			output.WriteString(fmt.Sprintf("--- PASS: %s\n", test.Name))
		}
	}

	if failedTests {
		output.WriteString(fmt.Sprintf("FAIL\t%s\n", time.Since(start).Round(time.Microsecond)))
	} else {
		output.WriteString(fmt.Sprintf("ok\t%s\n", time.Since(start).Round(time.Microsecond)))
	}

	sendJSONResponse(w, map[string]string{"output": output.String()})
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Config string            `json:"config"`
		Input  string            `json:"input"`
		Env    map[string]string `json:"env"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	conf, err := compileStr(request.Config, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error compiling config: %v", err), http.StatusBadRequest)
		return
	}

	var cfg substation.Config
	if err := json.Unmarshal([]byte(conf), &cfg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	// Set up environment variables
	for key, value := range request.Env {
		os.Setenv(key, value)
	}

	sub, err := substation.New(r.Context(), cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating Substation instance: %v", err), http.StatusInternalServerError)
		return
	}

	msgs := []*message.Message{
		message.New().SetData([]byte(request.Input)),
		message.New().AsControl(),
	}

	result, err := sub.Transform(r.Context(), msgs...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error transforming messages: %v", err), http.StatusInternalServerError)
		return
	}

	var output []string
	for _, msg := range result {
		if !msg.HasFlag(message.IsControl) {
			output = append(output, string(msg.Data()))
		}
	}

	// Clean up environment variables after processing
	for key := range request.Env {
		os.Unsetenv(key)
	}

	sendJSONResponse(w, map[string]interface{}{"output": output})
}

func handleFmt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Println("Received /fmt request")
	if r.Method != http.MethodPost {
		log.Println("Method not allowed:", r.Method)
		sendJSONResponse(w, map[string]string{"error": "Method not allowed"})
		return
	}

	var input struct {
		Jsonnet string `json:"jsonnet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Error decoding request: %v", err)
		log.Printf("Request body: %s", getRequestBody(r))
		sendJSONResponse(w, map[string]string{"error": fmt.Sprintf("Error decoding request: %v", err)})
		return
	}

	log.Printf("Received Jsonnet content: %s", input.Jsonnet)

	log.Println("Formatting Jsonnet...")
	formatted, err := formatter.Format("", input.Jsonnet, formatter.DefaultOptions())
	if err != nil {
		log.Printf("Error formatting Jsonnet: %v", err)
		sendJSONResponse(w, map[string]string{"error": fmt.Sprintf("Error formatting Jsonnet: %v", err)})
		return
	}

	sendJSONResponse(w, map[string]interface{}{"config": formatted})
}

func getRequestBody(r *http.Request) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Sprintf("Error reading body: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

// Add a new handler for sharing
func handleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Config string `json:"config"`
		Input  string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Combine and encode the data
	combined := request.Config + "{substation-separator}" + request.Input + "{substation-separator}"
	encoded := base64.URLEncoding.EncodeToString([]byte(combined))

	// Create the shareable URL
	shareURL := url.URL{
		Path:     "/",
		RawQuery: "share=" + encoded,
	}

	sendJSONResponse(w, map[string]string{"url": shareURL.String()})
}
