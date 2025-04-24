package helpers

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	// Make sure the debug package is imported correctly based on your project structure
	"github.com/jabafett/quill/internal/utils/debug"
)

// SuggestionGroup represents a group of files that should be committed together
type SuggestionGroup struct {
	ID          string   // Unique identifier for the group
	Description string   // Description of the group
	Files       []string // Files in the group
	Message     string   // Suggested commit message
	ShouldStage bool     // Whether the files should be staged
}

// ErrNoChanges is returned when there are no changes to suggest groupings for
type ErrNoChanges struct{}

func (e ErrNoChanges) Error() string {
	return "no changes found to suggest groupings for"
}

// ParseSuggestionResponse parses the AI response into structured suggestions
func ParseSuggestionResponse(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
	// First, try to parse as XML
	xmlGroups := parseXMLResponse(response, stagedFiles, unstagedFiles)
	if len(xmlGroups) > 0 {
		debug.Log("ParseSuggestionResponse: Successfully parsed %d groups using XML.", len(xmlGroups))
		return xmlGroups
	}
	debug.Log("ParseSuggestionResponse: XML parsing returned 0 groups, falling back to regex.")
	// Fallback to regex parsing for backward compatibility
	return parseRegexResponse(response, stagedFiles, unstagedFiles)
}

// parseXMLResponse parses XML-formatted AI responses with added debugging
func parseXMLResponse(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
	debug.Log("parseXMLResponse: Starting XML parsing.")
	var groups []SuggestionGroup

	// Extract XML content from the response (it might be surrounded by other text)
	xmlPattern := regexp.MustCompile(`<suggestions>(?s).*?</suggestions>`)
	xmlMatch := xmlPattern.FindString(response)
	if xmlMatch == "" {
		debug.Log("parseXMLResponse: No XML <suggestions> block found in the response.")
		return groups // No XML found
	}
	debug.Log("parseXMLResponse: Found XML block:\n%s", xmlMatch)

	// Define XML structure locally within the function
	type File struct {
		Value string `xml:",chardata"`
	}
	type Commit struct {
		Header string `xml:"header"`
		Body   string `xml:"body"`
		Footer string `xml:"footer"`
	}
	type Group struct {
		Description string `xml:"description"`
		Files       struct {
			File []File `xml:"file"`
		} `xml:"files"`
		Message string `xml:"message"` // For backward compatibility
		Commit  Commit `xml:"commit"`
	}
	type Suggestions struct {
		Groups []Group `xml:"group"`
	}

	// Parse the XML
	var suggestions Suggestions
	err := xml.Unmarshal([]byte(xmlMatch), &suggestions)
	if err != nil {
		debug.Log("parseXMLResponse: XML unmarshal error: %v", err)
		return groups // XML parsing failed
	}
	debug.Log("parseXMLResponse: Successfully unmarshalled %d groups from XML.", len(suggestions.Groups))

	// Combine staged and unstaged for easier validation checking later
	// This avoids creating it repeatedly inside the loop
	allKnownFiles := make(map[string]struct{}, len(stagedFiles)+len(unstagedFiles))
	for _, f := range stagedFiles {
		allKnownFiles[f] = struct{}{}
	}
	for _, f := range unstagedFiles {
		allKnownFiles[f] = struct{}{}
	}
	debug.Log("  All known files map for validation created with %d entries.", len(allKnownFiles))

	// Convert to SuggestionGroup objects
	for i, xmlGroup := range suggestions.Groups {
		debug.Log("parseXMLResponse: Processing XML Group %d: '%s'", i+1, xmlGroup.Description)
		var filesFromXML []string
		for _, file := range xmlGroup.Files.File {
			// Clean up potential surrounding quotes or backticks from AI output
			cleanFile := strings.Trim(strings.TrimSpace(file.Value), `"'`)
			if cleanFile != "" { // Avoid adding empty strings
				filesFromXML = append(filesFromXML, cleanFile)
			}
		}
		debug.Log("  Files listed in XML: %v", filesFromXML)

		// Validate files against the combined known files map
		validatedFiles := make([]string, 0, len(filesFromXML))
		for _, file := range filesFromXML {
			if _, exists := allKnownFiles[file]; exists {
				validatedFiles = append(validatedFiles, file)
			} else {
				debug.Log("  File '%s' from XML group not found in known staged/unstaged files.", file)
			}
		}
		debug.Log("  Validated Files for this group: %v", validatedFiles)

		// Only create a group if we have valid files associated with it
		if len(validatedFiles) > 0 {
			debug.Log("  Group has validated files, proceeding to create SuggestionGroup.")
			// Determine if files should be staged (based on validated files)
			// A group should be staged if *any* of its validated files are currently unstaged.
			shouldStage := false
			for _, file := range validatedFiles {
				// Check only against the original unstaged list
				if contains(unstagedFiles, file) { // Use the original unstagedFiles list here
					shouldStage = true
					debug.Log("  File '%s' is unstaged, marking group for staging.", file)
					break // Found one unstaged file, no need to check further
				}
			}
			debug.Log("  ShouldStage determined as: %v", shouldStage)

			// Construct the commit message from the structured parts
			var message string
			if xmlGroup.Commit.Header != "" {
				// Use the structured commit format
				message = xmlGroup.Commit.Header
				if xmlGroup.Commit.Body != "" {
					message += "\n\n" + xmlGroup.Commit.Body
				}
				if xmlGroup.Commit.Footer != "" {
					message += "\n\n" + xmlGroup.Commit.Footer
				}
				debug.Log("  Constructed message from <commit> tag.")
			} else if xmlGroup.Message != "" {
				// Fallback to the legacy message field
				message = xmlGroup.Message
				debug.Log("  Constructed message from legacy <message> tag.")
			} else {
				debug.Log("  Warning: No commit message (<commit> or <message>) found in XML group.")
			}
			debug.Log("  Final Message: %s", message)

			// Create the suggestion group
			group := SuggestionGroup{
				// ID is assigned later in the provider
				Description: xmlGroup.Description,
				Files:       validatedFiles, // IMPORTANT: Use the validated files list
				Message:     message,
				ShouldStage: shouldStage,
			}
			groups = append(groups, group)
			debug.Log("  Group successfully added.")

		} else {
			debug.Log("  Group SKIPPED (No validated files found for this group).")
		}
	} // End of loop through XML groups

	debug.Log("parseXMLResponse: Finished processing. Returning %d validated groups.", len(groups))
	return groups
}

// parseRegexResponse parses AI responses using regex (legacy format)
// ... (keep this function as it was, it's the fallback)
func parseRegexResponse(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
	debug.Log("parseRegexResponse: Starting regex parsing.")
	var groups []SuggestionGroup

	// Pattern for finding groupings - more comprehensive
	groupPattern := regexp.MustCompile(`(?i)(?:Group|Grouping|Suggestion|Commit)\s*(?:group)?\s*\d*:?\s*([^\n]+)`)
	filePattern := regexp.MustCompile(`(?i)(?:Files?|Include|Changes):\s*([^\n]+(?:\n\s*[-*]\s*[^\n]+)*)`)
	messagePattern := regexp.MustCompile(`(?i)(?:Commit message|Message|Description|Commit):\s*([^\n]+)`)

	// Find all groupings
	groupMatches := groupPattern.FindAllStringSubmatchIndex(response, -1)
	debug.Log("parseRegexResponse: Found %d potential group matches via regex.", len(groupMatches))

	// Combine staged and unstaged for easier validation checking later
	allKnownFiles := make(map[string]struct{}, len(stagedFiles)+len(unstagedFiles))
	for _, f := range stagedFiles {
		allKnownFiles[f] = struct{}{}
	}
	for _, f := range unstagedFiles {
		allKnownFiles[f] = struct{}{}
	}

	// Process standard groupings
	for i, groupMatch := range groupMatches {
		if len(groupMatch) < 4 {
			debug.Log("parseRegexResponse: Skipping invalid regex group match index %d.", i)
			continue
		}

		// Extract the group description
		description := strings.TrimSpace(response[groupMatch[2]:groupMatch[3]])
		debug.Log("parseRegexResponse: Processing Regex Group %d: '%s'", i+1, description)

		// Determine the end of this group (either the start of the next group or the end of the response)
		groupEnd := len(response)
		if i+1 < len(groupMatches) {
			groupEnd = groupMatches[i+1][0]
		}

		// Extract the group content
		groupContent := response[groupMatch[0]:groupEnd]

		// Extract files
		var filesFromRegex []string
		fileMatches := filePattern.FindStringSubmatch(groupContent)
		if len(fileMatches) > 1 {
			fileList := fileMatches[1]
			debug.Log("  Files string from regex: %s", fileList)
			// Check if the file list contains bullet points
			if strings.Contains(fileList, "-") || strings.Contains(fileList, "*") {
				// Extract files from bullet points
				bulletPattern := regexp.MustCompile(`[-*]\s*([^\n,;]+)`)
				bulletMatches := bulletPattern.FindAllStringSubmatch(fileList, -1)
				for _, match := range bulletMatches {
					if len(match) > 1 {
						file := strings.Trim(strings.TrimSpace(match[1]), `"'`)
						if file != "" {
							filesFromRegex = append(filesFromRegex, file)
						}
					}
				}
			} else {
				// Split by commas, semicolons, or newlines
				for _, file := range regexp.MustCompile(`[,;\n]+`).Split(fileList, -1) {
					file := strings.Trim(strings.TrimSpace(file), `"'`)
					if file != "" {
						filesFromRegex = append(filesFromRegex, file)
					}
				}
			}
		}
		debug.Log("  Files extracted via regex: %v", filesFromRegex)

		// Extract commit message
		message := ""
		messageMatches := messagePattern.FindStringSubmatch(groupContent)
		if len(messageMatches) > 1 {
			message = strings.Trim(strings.TrimSpace(messageMatches[1]), `"'`)
		}
		debug.Log("  Message extracted via regex: %s", message)

		// Validate files against the repository
		validatedFiles := make([]string, 0, len(filesFromRegex))
		for _, file := range filesFromRegex {
			if _, exists := allKnownFiles[file]; exists {
				validatedFiles = append(validatedFiles, file)
			} else {
				debug.Log("  File '%s' from regex group not found in known staged/unstaged files.", file)
			}
		}
		debug.Log("  Validated Files for this regex group: %v", validatedFiles)

		// Only create a group if we have valid files
		if len(validatedFiles) > 0 {
			debug.Log("  Regex group has validated files, proceeding.")
			shouldStage := false
			for _, file := range validatedFiles {
				if contains(unstagedFiles, file) { // Check original unstaged list
					shouldStage = true
					debug.Log("  File '%s' is unstaged, marking regex group for staging.", file)
					break
				}
			}
			debug.Log("  ShouldStage determined as: %v", shouldStage)

			// Create the suggestion group
			group := SuggestionGroup{
				Description: description,
				Files:       validatedFiles,
				Message:     message,
				ShouldStage: shouldStage,
			}
			groups = append(groups, group)
			debug.Log("  Regex group successfully added.")
		} else {
			debug.Log("  Regex Group SKIPPED (No validated files found for this group).")
		}
	} // End of loop through regex groups

	// If no groups were found via specific regex patterns, try fallback
	if len(groups) == 0 {
		debug.Log("parseRegexResponse: No specific regex groups found, attempting fallback parsing.")
		groups = fallbackParsing(response, stagedFiles, unstagedFiles) // Pass lists for validation
	}

	debug.Log("parseRegexResponse: Finished regex processing. Returning %d groups.", len(groups))
	return groups
}

// fallbackParsing tries to extract file paths and commit messages from unstructured text
func fallbackParsing(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
	debug.Log("fallbackParsing: Starting fallback parsing.")
	var groups []SuggestionGroup

	// Combine for validation
	allKnownFiles := make(map[string]struct{}, len(stagedFiles)+len(unstagedFiles))
	for _, f := range stagedFiles {
		allKnownFiles[f] = struct{}{}
	}
	for _, f := range unstagedFiles {
		allKnownFiles[f] = struct{}{}
	}

	// Look for file paths with common extensions more broadly
	// This regex is simplified to find potential paths, validation is key
	fileExtPattern := regexp.MustCompile(`(?m)^\s*[-*]?\s*([\w./-]+\.[\w]+)`) // Look for lines starting with optional bullet/dash and a path-like string with an extension
	extMatches := fileExtPattern.FindAllStringSubmatch(response, -1)
	debug.Log("fallbackParsing: Found %d potential file matches.", len(extMatches))

	var files []string
	for _, match := range extMatches {
		if len(match) > 1 {
			potentialFile := strings.Trim(strings.TrimSpace(match[1]), `"'`)
			if _, exists := allKnownFiles[potentialFile]; exists && !contains(files, potentialFile) {
				files = append(files, potentialFile)
			}
		}
	}
	debug.Log("fallbackParsing: Validated files from fallback: %v", files)

	// Look for commit message (conventional commit format is a good heuristic)
	message := ""
	conventionalPattern := regexp.MustCompile(`(?im)^(feat|fix|docs|style|refactor|test|chore)(\([^)]+\))?:\s*(.+)`) // Case-insensitive, multiline
	conventionalMatches := conventionalPattern.FindStringSubmatch(response)
	if len(conventionalMatches) > 3 {
		scope := ""
		if conventionalMatches[2] != "" {
			scope = conventionalMatches[2] // Includes the parentheses already
		}
		message = conventionalMatches[1] + scope + ": " + strings.TrimSpace(conventionalMatches[3])
		debug.Log("fallbackParsing: Found conventional commit message: %s", message)
	} else {
		// Maybe just grab the first non-empty line as a potential message?
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "<") && !strings.HasPrefix(trimmedLine, "-") && !strings.HasPrefix(trimmedLine, "*") { // Avoid XML tags and list items
				message = trimmedLine
				debug.Log("fallbackParsing: Using first non-empty line as message: %s", message)
				break
			}
		}
	}

	// Only create a group if we found *something* useful (validated files or a message)
	if len(files) > 0 || message != "" {
		debug.Log("fallbackParsing: Creating fallback group.")
		shouldStage := false
		for _, file := range files { // Iterate through validated files
			if contains(unstagedFiles, file) { // Check original unstaged list
				shouldStage = true
				debug.Log("  File '%s' is unstaged, marking fallback group for staging.", file)
				break
			}
		}
		debug.Log("  ShouldStage determined as: %v", shouldStage)

		groups = append(groups, SuggestionGroup{
			Description: "Suggested changes (fallback)", // Generic description
			Files:       files,                          // Use validated files
			Message:     message,
			ShouldStage: shouldStage,
		})
		debug.Log("fallbackParsing: Fallback group added.")
	} else {
		debug.Log("fallbackParsing: No useful information found for fallback group.")
	}

	return groups
}

// contains checks if a string is in a slice (consider using the map for O(1) lookup if performance matters)
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SplitLines splits a string into lines
func SplitLines(s string) []string {
	// Normalize line endings before splitting
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(strings.TrimSpace(normalized), "\n")
}

// ExecuteCommand executes a shell command and returns its output
func ExecuteCommand(cmd string) (string, error) {
	// Split the command string into command and arguments
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil // Or return an error?
	}
	command := parts[0]
	args := parts[1:]

	// Execute the command
	// Use CombinedOutput to capture both stdout and stderr for better error diagnosis
	outputBytes, err := exec.Command(command, args...).CombinedOutput()
	output := string(outputBytes)
	if err != nil {
		// Include the command output in the error message if execution failed
		return output, fmt.Errorf("failed to execute command '%s': %w\nOutput:\n%s", cmd, err, output)
	}

	return output, nil
}
