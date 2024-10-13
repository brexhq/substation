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
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
	"github.com/google/go-jsonnet"
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
	mux.HandleFunc("/examples", handleExamples)

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
		DefaultConfig: demoConf,
		DefaultInput:  demoEvt,
	}
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			output = append(output, gjson.Get(string(msg.Data()), "@this|@pretty").String())
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"output": output,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	examples := map[string]struct {
		Config string `json:"config"`
		Input  string `json:"input"`
	}{
		"stringConversion": {
			Config: `{
  transforms: [
    sub.tf.time.from.string({ obj: { source_key: 'time', target_key: 'time' }, format: '2006-01-02T15:04:05.000Z' }),
    sub.tf.time.to.string({ obj: { source_key: 'time', target_key: 'time' }, format: '2006-01-02T15:04:05' }),
  ],
}`,
			Input: `{"time":"2024-01-01T01:02:03.123Z"}`,
		},
		"numberClamp": {
			Config: `{
  transforms: [
    sub.tf.number.maximum({ value: 0 }),
    sub.tf.number.minimum({ value: 100 }),
  ],
}`,
			Input: `-1
101
50`,
		},
		"arrayFlatten": {
			Config: `{
  transforms: [
    sub.tf.obj.cp({ object: { source_key: 'a|@flatten', target_key: 'a' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
  ],
}`,
			Input: `{"a":[1,2,[3,4]]}`,
		},
	}

	if err := json.NewEncoder(w).Encode(examples); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
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
            max-width: 90vw;
            margin: 0 auto;
            padding: 40px;
            background-color: #f9f9f9;
            color: var(--text-color);
            display: grid;
            grid-template-rows: auto 1fr;
            height: 100vh;
            box-sizing: border-box;
        }

        header {
            margin-bottom: 40px;
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
        }

        .title {
            gap: 16px;
        }

        .title-container {
            position: relative;
            padding-bottom: 1em;
        }

        h1 {
            font-size: 48px;
            color: #212121;
            font-weight: 800;
            margin-bottom: 8px;
            word-wrap: break-word;
        }

        .playground-label {
            font-weight: 300;
            font-family: 'Inter', sans-serif;
            color: var(--secondary-color);
            opacity: 0.5;
        }

        h2 {
            font-size: 24px;
            color: #202020;
            font-weight: 700;
            margin-top: 0;
            margin-bottom: 4px;
        }

        h3 {
            font-weight: 500;
            color: #666666;
            font-size: 18px;
            margin-top: 0;
        }

        main {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 40px;
            height: 100%;
        }

        .left-column,
        .right-column {
            display: flex;
            flex-direction: column;
            gap: 20px;
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
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 0 0 1px var(--border-color), 0 2px 4px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }

        .button-container {
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: 5px;
        }

        .action-row {
            display: flex;
            flex-direction: row;
            align-items: center;
            gap: 10px;
        }

        .select-container {
            display: flex;
            flex-direction: column;
        }

        .select-container select {
            height: 40px;
            box-sizing: border-box;
            padding: 0 10px;
        }

        .subtext {
            font-size: 12px;
            color: var(--secondary-color);
            margin: 5px 0 8px 0;  // Added bottom margin
        }

        button {
            padding: 0 48px;
            height: 40px;
            color: white;
            border: none;
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

        .secondary-link {
            color: var(--secondary-color);
            text-decoration: none;
            font-weight: 500;
        }

        button:active {
            transform: translateY(1px);
        }

        @media (max-width: 1200px) {
            body {
                max-width: 95vw;
            }

            main {
                grid-template-columns: 1fr;
            }

            h1 {
                font-size: 36px;
            }

            h3 {
                font-size: 16px;
            }
        }


        .nav-bar {
            position: absolute;
            top: 20px;
            right: 40px;
            display: flex;
            gap: 20px;
        }

        .nav-link {
            color: var(--secondary-color);
            text-decoration: none;
            font-size: 24px;
            transition: color 0.3s ease;
        }

        .nav-link:hover {
            color: var(--secondary-hover-color);
        }

        select {
            padding: 10px;
            font-size: 16px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            background-color: #ffffff;
            color: var(--text-color);
            cursor: pointer;
        }

        select:hover {
            border-color: var(--primary-color);
        }

        .logo {
            height: 36px;
            width: auto;
        }

        .title {
            display: flex;
            align-items: center;
        }

        .logo-container {
            position: absolute;
            top: 20px;
            left: 40px;
        }
    </style>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.30.1/min/vs/loader.min.js"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <link rel="icon" type="image/png" href="https://files.readme.io/2f32047-small-substation_logo.png">
</head>

<body>
    <header>
        <div class="nav-bar">
            <a href="https://substation.readme.io/docs/overview" target="_blank" class="nav-link" title="Documentation">
                <i class="fas fa-book"></i>
            </a>
            <a href="https://github.com/brexhq/substation" target="_blank" class="nav-link" title="GitHub">
                <i class="fab fa-github"></i>
            </a>
        </div>
        <div>
            <div class="title-container">
                <h1 class="title">
                    <span>Substation</span>
                    <span class="playground-label">Playground</span>
                </h1>
                <h3>A toolkit for routing, normalizing, and enriching security event and audit logs.</h3>
            </div>
            <div class="button-container">
                <div class="action-row">
                    <div class="select-container">
                        <select id="exampleSelector" onchange="loadExample()">
                            <option value="stringConversion">String Conversion</option>
                            <option value="numberClamp">Number Clamp</option>
                            <option value="arrayFlatten">Array Flatten</option>
                        </select>
                    </div>
                    <button class="primary-button" onclick="runSubstation()">Run</button>
                </div>
                <p class="subtext">Select an example to get started or create your own configuration.</p>
            </div>
        </div>
    </header>
    <main>
        <section class="left-column">
            <div class="editor-section">
                <h2>Configuration</h2>
                <p class="subtext">Configure the transformations to be applied to the input event.</p>
                <div class="editor-container" id="config"></div>
            </div>
        </section>
        <section class="right-column">
            <div class="editor-section">
                <h2>Input</h2>
                <p class="subtext">Paste the JSON event to be processed by Substation here.</p>
                <div class="editor-container" id="input"></div>
            </div>
            <div class="editor-section">
                <h2>Output</h2>
                <p class="subtext">The processed event will appear here after running.</p>
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
                });
            }


            configEditor = createEditor('config', 'jsonnet', "");
            inputEditor = createEditor('input', 'json', "");
            outputEditor = createEditor('output', 'json', '// Output will appear here');

            // Fetch examples from the API and set default example
            fetch('/examples')
                .then(response => response.json())
                .then(data => {
                    examples = data;
                    // Set the default example to "stringConversion"
                    document.getElementById('exampleSelector').value = 'stringConversion';
                    loadExample();
                })
                .catch(error => console.error('Error fetching examples:', error));
        });

        function runSubstation() {
            fetch('/run', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    config: configEditor.getValue(),
                    input: inputEditor.getValue()
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

        function loadExample() {
            const example = document.getElementById('exampleSelector').value;
            if (example in examples) {
                configEditor.setValue(examples[example].config);
                inputEditor.setValue(examples[example].input);
                outputEditor.setValue('// Output will appear here');
            } else if (example === '') {
                configEditor.setValue("");
                inputEditor.setValue("");
                outputEditor.setValue('// Output will appear here');
            }
        }
    </script>
</body>

</html>
`
