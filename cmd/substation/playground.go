package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/formatter"
)

func init() {
	rootCmd.AddCommand(playgroundCmd)
}

var playgroundCmd = &cobra.Command{
	Use:   "playground",
	Short: "start playground",
	Long:  `'substation playground' starts a local HTTP server for testing Substation configurations.`,
	RunE:  runPlayground,
}

func runPlayground(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/demo", handleDemo)
	mux.HandleFunc("/fmt", handleFmt)

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
	}{
		DefaultConfig: "",
		DefaultInput:  "",
	}

	tmpl := template.Must(template.New("index").Parse(indexHTML))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cleanedDemoconf := strings.ReplaceAll(demoConf, "local sub = import '../../substation.libsonnet';\n\n", "")

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"config": cleanedDemoconf,
		"input":  demoEvt,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func handleRun(w http.ResponseWriter, r *http.Request) {
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

	combinedConfig := fmt.Sprintf(`local sub = %s;

%s`, substation.Libsonnet, request.Config)

	vm := jsonnet.MakeVM()
	jsonString, err := vm.EvaluateAnonymousSnippet("", combinedConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error evaluating Jsonnet: %v", err), http.StatusBadRequest)
		return
	}

	var cfg substation.Config
	if err := json.Unmarshal([]byte(jsonString), &cfg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
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
		if !msg.IsControl() {
			output = append(output, string(msg.Data()))
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"output": output,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func handleFmt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Jsonnet string `json:"jsonnet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}

	formatted, err := formatter.Format("", input.Jsonnet, formatter.DefaultOptions())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error formatting Jsonnet: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"formatted": formatted}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

const indexHTML = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Substation | Playground</title>
    <meta name="description" content="A toolkit for routing, normalizing, and enriching security event and audit logs.">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');

        :root {
            --primary-color: #F46A35;
            --primary-hover-color: #E55A25;
            --text-color: #1c1c1c;
            --border-color: #D9D9D9;
            --secondary-color: #6c757d;
            --secondary-hover-color: #5a6268;
        }

        body {
            font-family: 'Inter', sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f9f9f9;
            color: var(--text-color);
            display: flex;
            flex-direction: column;
            min-height: 95vh;
            box-sizing: border-box;
        }

        .content-wrapper {
            margin: 0 auto;
            padding: 0 40px;
            width: 100%;
            box-sizing: border-box;
        }

        .nav-bar {
            background-color: #ffffff;
            padding: 10px 0;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .nav-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .title {
            font-size: 20px;
            font-weight: 800;
            color: #212121;
        }

        .playground-label {
            font-weight: 300;
            color: var(--secondary-color);
            opacity: 0.5;
        }

        .nav-links {
            display: flex;
            gap: 20px;
        }

        .nav-link {
            color: var(--secondary-color);
            text-decoration: none;
            font-size: 20px;
            transition: color 0.3s ease;
        }

        .nav-link:hover {
            color: var(--secondary-hover-color);
        }

        .action-section {
            padding: 20px 0;
            background-color: #f0f0f04e;
            border-bottom: 1px solid var(--border-color);
        }

        .button-container {
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: 10px;
        }

        .action-row {
            display: flex;
            flex-direction: row;
            align-items: center;
            gap: 10px;
        }

        main {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 40px;
            flex-grow: 1;
            overflow: hidden;
            padding: 20px 0;
        }

        .left-column,
        .right-column {
            display: flex;
            flex-direction: column;
            gap: 18px;
            overflow: hidden;
        }

        .right-column {
            grid-template-rows: 1fr 1fr;
        }

        .editor-section {
            display: flex;
            flex-direction: column;
            flex-grow: 1;
        }

        .editor-container {
            flex-grow: 1;
            background-color: #1e1e1e;
            border-radius: 8px;
            padding: 16px 0px;
            box-shadow: 0 0 0 1px var(--border-color), 0 2px 4px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }

        .subtext {
            font-size: 12px;
            color: var(--secondary-color);
            margin: 5px 0 8px 0; 
        }

        button {
            padding: 0 24px;
            height: 36px;
            color: white;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            cursor: pointer;
            font-family: 'Inter', sans-serif;
            font-weight: 600;
            font-size: 16px;
            transition: background-color 0.3s ease, transform 0.1s ease;
            box-sizing: border-box;
        }

        .primary-button {
            background-color: var(--primary-color);
        }

        .primary-button:hover {
            background-color: var(--primary-hover-color);
        }

        .secondary-button {
            background-color: #EDEFEE;
            color: #323333;
        }

        .secondary-button:hover {
            background-color: #D9DBD9;
        }

        button:active {
            transform: translateY(1px);
        }

        .examples-link {
            color: var(--primary-color);
            text-decoration: none;
        }

        .examples-link:hover {
            text-decoration: underline;
        }

        @media (max-width: 1200px) {
            main {
                grid-template-columns: 1fr;
            }
            .content-wrapper {
                padding: 0 20px;
            }
        }

        h2 {
            margin: 24px 0px 0px 0px
        }

        .editor-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 4px;
        }

        .mode-selector {
            font-size: 14px;
            padding: 5px;
            border-radius: 4px;
            border: 1px solid var(--border-color);
            background-color: #ffffff;
            color: var(--text-color);
        }
    </style>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.30.1/min/vs/loader.min.js"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <link rel="icon" type="image/png" href="https://files.readme.io/2f32047-small-substation_logo.png">
</head>

<body>
    <nav class="nav-bar">
        <div class="content-wrapper">
            <div class="nav-content">
                <div class="title">
                    Substation <span class="playground-label">Playground</span>
                </div>
                <div class="nav-links">
                    <a href="https://substation.readme.io/docs/overview" target="_blank" class="nav-link" title="Documentation">
                        <i class="fas fa-book"></i>
                    </a>
                    <a href="https://github.com/brexhq/substation" target="_blank" class="nav-link" title="GitHub">
                        <i class="fab fa-github"></i>
                    </a>
                </div>
            </div>
        </div>
    </nav>
    <section class="action-section">
        <div class="content-wrapper">
            <div class="button-container">
                <div class="action-row">
                    <button class="primary-button" onclick="runSubstation()">Run</button>
                    <button class="secondary-button" onclick="testSubstation()">Test</button>
                    <button class="secondary-button" onclick="demoSubstation()">Demo</button>
                    <button class="secondary-button" onclick="formatJsonnet()">Format</button>
                </div>
                <p class="subtext">
                    Run your configuration, test it, or try a demo. 
                    <a href="https://github.com/brexhq/substation/tree/main/examples" target="_blank" class="examples-link">View examples</a>
                </p>
            </div>
        </div>
    </section>
    <main class="content-wrapper">
        <section class="left-column">
            <div class="editor-section">
                <div class="editor-header">
                    <h2>Configuration</h2>
                </div>
                <p class="subtext">Configure the transformations to be applied to the input event.</p>
                <div class="editor-container" id="config"></div>
            </div>
        </section>
        <section class="right-column">
            <div class="editor-section">
                <div class="editor-header">
                    <h2>Input</h2>
                    <select class="mode-selector" id="inputModeSelector" onchange="changeEditorMode('input')">
                        <option value="text">Text</option>
                        <option value="json">JSON</option>
                    </select>
                </div>
                <p class="subtext">Paste the message data to be processed by Substation here.</p>
                <div class="editor-container" id="input"></div>
            </div>
            <div class="editor-section">
                <div class="editor-header">
                    <h2>Output</h2>
                    <select class="mode-selector" id="outputModeSelector" onchange="changeEditorMode('output')">
                        <option value="text">Text</option>
                        <option value="json">JSON</option>
                    </select>
                </div>
                <p class="subtext">The processed message data will appear here after running.</p>
                <div class="editor-container" id="output"></div>
            </div>
        </section>
    </main>

    <script>
        let configEditor, inputEditor, outputEditor;
        let examples = {};

        require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.30.1/min/vs' } });

        require(['vs/editor/editor.main'], function () {
            function createEditor(elementId, language, value) {
                return monaco.editor.create(document.getElementById(elementId), {
                    value: value,
                    language: language,
                    theme: 'vs-dark',
                    automaticLayout: true,
                    minimap: { enabled: false },
                    scrollBeyondLastLine: false,
                    lineNumbers: 'on',
                    roundedSelection: false,
                    readOnly: elementId === 'output',
                    renderLineHighlight: 'none',
                    wordWrap: 'on', 
                });
            }

            configEditor = createEditor('config', 'jsonnet', "");
            inputEditor = createEditor('input', 'text', "");
            outputEditor = createEditor('output', 'text', '// Processed message data will appear here');
        });

        function changeEditorMode(editorId) {
            const editor = editorId === 'input' ? inputEditor : outputEditor;
            const id = editorId + "ModeSelector";
            const selector = document.getElementById(id);
            const newModel = monaco.editor.createModel(editor.getValue(), selector.value);
            editor.setModel(newModel);
        }

        function runSubstation() {
            fetch('/run', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    config: configEditor.getValue(),
                    input: inputEditor.getValue(),
                })
            })
                .then(response => response.json())
                .then(data => {
                    outputEditor.setValue(data.output.join('\n'));
                })
                .catch(error => {
                    outputEditor.setValue('Error: ' + error);
                });
        }

        function formatJsonnet() {
            fetch('/fmt', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ jsonnet: configEditor.getValue() })
            })
                .then(response => response.json())
                .then(data => {
                    configEditor.setValue(data.formatted);
                })
                .catch(error => console.error('Error formatting Jsonnet:', error));
        }

        function testSubstation() {
            fetch('/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    config: configEditor.getValue(),
                    input: inputEditor.getValue(),
                })
            })
            .then(response => response.json())
            .then(data => {
                outputEditor.setValue(JSON.stringify(data, null, 2));
            })
            .catch(error => {
                outputEditor.setValue('Error: ' + error);
            });
        }

        function demoSubstation() {
            fetch('/demo')
            .then(response => response.json())
            .then(data => {
                configEditor.setValue(data.config);
                inputEditor.setValue(data.input);
                outputEditor.setValue('// Run the demo to see the output');
            })
            .catch(error => {
                console.error('Error fetching demo:', error);
            });
        }
    </script>
</body>

</html>
`
