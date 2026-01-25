package thinci

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// Executor handles the execution of CI jobs locally
type Executor struct {
	verbose bool
	dryRun  bool
}

// NewExecutor creates a new executor
func NewExecutor(verbose, dryRun bool) *Executor {
	return &Executor{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// ExecuteJob runs a single job from a plan
func (e *Executor) ExecuteJob(job Job) error {
	jobID := job.GetID()
	action := job.GetAction()
	component := job.GetComponent()
	
	e.logSection(fmt.Sprintf("Executing Job: %s", jobID))
	e.logInfo(fmt.Sprintf("Component: %s", component))
	e.logInfo(fmt.Sprintf("Action: %s", action))
	
	startTime := time.Now()
	
	// Extract job fields
	preSteps := e.extractSteps(job, "preSteps")
	commands := e.extractCommands(job, "commands")
	postSteps := e.extractSteps(job, "postSteps")
	inputs := e.extractInputs(job, "inputs")
	
	// Create template context for variable substitution
	context := e.buildTemplateContext(job, inputs)
	
	// Execute pre-steps
	if len(preSteps) > 0 {
		e.logSection("Pre-Steps")
		if err := e.executeSteps(preSteps, context); err != nil {
			return fmt.Errorf("pre-steps failed: %w", err)
		}
	}
	
	// Execute main commands
	if len(commands) > 0 {
		e.logSection("Main Commands")
		if err := e.executeCommands(commands, context); err != nil {
			return fmt.Errorf("commands failed: %w", err)
		}
	}
	
	// Execute post-steps
	if len(postSteps) > 0 {
		e.logSection("Post-Steps")
		if err := e.executeSteps(postSteps, context); err != nil {
			return fmt.Errorf("post-steps failed: %w", err)
		}
	}
	
	duration := time.Since(startTime)
	e.logSuccess(fmt.Sprintf("Job completed successfully in %s", duration.Round(time.Millisecond)))
	
	return nil
}

// buildTemplateContext creates a map for template variable substitution
func (e *Executor) buildTemplateContext(job Job, inputs map[string]any) map[string]string {
	context := make(map[string]string)
	
	// Add core job fields
	context["id"] = job.GetID()
	context["component"] = job.GetComponent()
	context["provider"] = job.GetProvider()
	context["action"] = job.GetAction()
	
	// Add all inputs as strings for template resolution
	for key, value := range inputs {
		context[key] = fmt.Sprintf("%v", value)
	}
	
	// Add common defaults
	if _, exists := context["releaseName"]; !exists {
		context["releaseName"] = job.GetComponent()
	}
	if _, exists := context["namespace"]; !exists {
		context["namespace"] = "default"
	}
	if _, exists := context["chartPath"]; !exists {
		context["chartPath"] = "."
	}
	if _, exists := context["valuesPath"]; !exists {
		context["valuesPath"] = "values.yaml"
	}
	if _, exists := context["timeout"]; !exists {
		context["timeout"] = "10m"
	}
	
	return context
}

// extractSteps extracts ActionStep array from job
func (e *Executor) extractSteps(job Job, fieldName string) []ActionStep {
	steps := []ActionStep{}
	
	if stepsRaw, ok := job[fieldName]; ok {
		switch v := stepsRaw.(type) {
		case []ActionStep:
			steps = v
		case []interface{}:
			for _, stepRaw := range v {
				if stepMap, ok := stepRaw.(map[string]interface{}); ok {
					step := ActionStep{
						Name:    getString(stepMap, "name"),
						Command: getString(stepMap, "command"),
					}
					if inputsMap, ok := stepMap["inputs"].(map[string]interface{}); ok {
						step.Inputs = make(map[string]any)
						for k, v := range inputsMap {
							step.Inputs[k] = v
						}
					}
					steps = append(steps, step)
				}
			}
		}
	}
	
	return steps
}

// extractCommands extracts command array from job
func (e *Executor) extractCommands(job Job, fieldName string) []string {
	commands := []string{}
	
	if commandsRaw, ok := job[fieldName]; ok {
		switch v := commandsRaw.(type) {
		case []string:
			commands = v
		case []interface{}:
			for _, cmd := range v {
				if cmdStr, ok := cmd.(string); ok {
					commands = append(commands, cmdStr)
				}
			}
		}
	}
	
	return commands
}

// extractInputs extracts inputs map from job
func (e *Executor) extractInputs(job Job, fieldName string) map[string]any {
	inputs := make(map[string]any)
	
	if inputsRaw, ok := job[fieldName]; ok {
		if inputsMap, ok := inputsRaw.(map[string]interface{}); ok {
			inputs = inputsMap
		}
	}
	
	return inputs
}

// executeSteps executes a list of action steps
func (e *Executor) executeSteps(steps []ActionStep, context map[string]string) error {
	for i, step := range steps {
		e.logStep(i+1, step.Name)
		
		// Resolve template variables in command
		command, err := e.resolveTemplate(step.Command, context)
		if err != nil {
			return fmt.Errorf("failed to resolve template in step '%s': %w", step.Name, err)
		}
		
		e.logCommand(command)
		
		if !e.dryRun {
			if err := e.runCommand(command); err != nil {
				return fmt.Errorf("step '%s' failed: %w", step.Name, err)
			}
		} else {
			e.logInfo("[DRY RUN] Command skipped")
		}
	}
	
	return nil
}

// executeCommands executes a list of commands
func (e *Executor) executeCommands(commands []string, context map[string]string) error {
	for i, cmdTemplate := range commands {
		e.logStep(i+1, fmt.Sprintf("Command %d", i+1))
		
		// Resolve template variables in command
		command, err := e.resolveTemplate(cmdTemplate, context)
		if err != nil {
			return fmt.Errorf("failed to resolve template in command: %w", err)
		}
		
		e.logCommand(command)
		
		if !e.dryRun {
			if err := e.runCommand(command); err != nil {
				return fmt.Errorf("command failed: %w", err)
			}
		} else {
			e.logInfo("[DRY RUN] Command skipped")
		}
	}
	
	return nil
}

// resolveTemplate resolves Go template variables in a string
func (e *Executor) resolveTemplate(templateStr string, context map[string]string) (string, error) {
	tmpl, err := template.New("command").Parse(templateStr)
	if err != nil {
		return "", err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// runCommand executes a shell command and streams output
func (e *Executor) runCommand(cmdStr string) error {
	// Use shell to execute command (handles pipes, redirects, etc.)
	cmd := exec.Command("sh", "-c", cmdStr)
	
	// Set environment variables
	cmd.Env = os.Environ()
	
	// Set up output handling
	if e.verbose {
		cmd.Stdout = &prefixWriter{prefix: "  │ ", writer: os.Stdout}
		cmd.Stderr = &prefixWriter{prefix: "  │ ", writer: os.Stderr}
		
		// Run the command (verbose mode)
		if err := cmd.Run(); err != nil {
			e.logError(fmt.Sprintf("Command failed with exit code %d", cmd.ProcessState.ExitCode()))
			e.logError(fmt.Sprintf("Command was: %s", cmdStr))
			return fmt.Errorf("command failed: %w", err)
		}
	} else {
		// Capture but don't display unless there's an error
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		// Run the command
		if err := cmd.Run(); err != nil {
			// Show detailed error information
			e.logError(fmt.Sprintf("Command failed with exit code %d", cmd.ProcessState.ExitCode()))
			e.logError(fmt.Sprintf("Command was: %s", cmdStr))
			
			// Show output on error if not verbose
			if stderr.Len() > 0 {
				fmt.Fprintf(os.Stderr, "\n  ┌─ Error Output:\n")
				for _, line := range strings.Split(strings.TrimSpace(stderr.String()), "\n") {
					fmt.Fprintf(os.Stderr, "  │ %s\n", line)
				}
				fmt.Fprintf(os.Stderr, "  └─\n")
			}
			if stdout.Len() > 0 {
				fmt.Fprintf(os.Stdout, "\n  ┌─ Standard Output:\n")
				for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
					fmt.Fprintf(os.Stdout, "  │ %s\n", line)
				}
				fmt.Fprintf(os.Stdout, "  └─\n")
			}
			
			return fmt.Errorf("command failed with exit code %d: %w", cmd.ProcessState.ExitCode(), err)
		}
	}
	
	return nil
}

// Logging helpers

func (e *Executor) logSection(message string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  %s\n", message)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (e *Executor) logStep(num int, name string) {
	fmt.Printf("\n  ▸ Step %d: %s\n", num, name)
}

func (e *Executor) logCommand(cmd string) {
	if e.verbose {
		fmt.Printf("  ├─ Command: %s\n", cmd)
		fmt.Println("  ├─ Output:")
	}
}

func (e *Executor) logInfo(message string) {
	fmt.Printf("  ℹ %s\n", message)
}

func (e *Executor) logSuccess(message string) {
	fmt.Printf("\n  ✓ %s\n", message)
}

func (e *Executor) logError(message string) {
	fmt.Fprintf(os.Stderr, "\n  ✗ %s\n", message)
}

// prefixWriter adds a prefix to each line of output
type prefixWriter struct {
	prefix string
	writer *os.File
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			break
		}
		fmt.Fprintf(pw.writer, "%s%s\n", pw.prefix, line)
	}
	return len(p), nil
}

// Helper function to get string from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
