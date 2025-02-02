package handlers

import (
	"bytes"
	"dev-team/internal/state"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

func HandleGeminiModels(w http.ResponseWriter, r *http.Request) {
	if state.State.GenAI == nil {
		http.Error(w, "AI provider not initialized", http.StatusInternalServerError)
		return
	}
	models := state.State.GenAI.Models()

	json.NewEncoder(w).Encode(models)
}

func HandleCodeGeneration(w http.ResponseWriter, r *http.Request) {
    // Assuming code generation logic is here
    // For now, we'll just validate the code
	validationOutput, err := validateCode()
	if err != nil {
        w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Code validation failed: %v", err)
        return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", validationOutput)

}

func validateCode() (string, error) {
	var outbuf, errbuf bytes.Buffer

	// go fmt ./...
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("go fmt failed: %s\n%s", err.Error(), errbuf.String())
	}

	// go mod tidy
	outbuf.Reset()
	errbuf.Reset()
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
        return "", fmt.Errorf("go mod tidy failed: %s\n%s", err.Error(), errbuf.String())
	}

	// golangci-lint run
	outbuf.Reset()
	errbuf.Reset()
	cmd = exec.Command("./bin/golangci-lint", "run")
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("golangci-lint failed: %s\n%s", err.Error(), errbuf.String())
	}

	// go test ./...
	outbuf.Reset()
	errbuf.Reset()
	cmd = exec.Command("go", "test", "./...")
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("go test failed: %s\n%s", err.Error(), errbuf.String())
	}

	return "Code validation passed", nil
}
