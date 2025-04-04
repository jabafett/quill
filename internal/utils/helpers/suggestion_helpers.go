package helpers

import (
"encoding/xml"
"os/exec"
"regexp"
"strings"
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
    var groups []SuggestionGroup

    // First, try to parse as XML
    xmlGroups := parseXMLResponse(response, stagedFiles, unstagedFiles)
    if len(xmlGroups) > 0 {
        return xmlGroups
    }

    // Fallback to regex parsing for backward compatibility
    return parseRegexResponse(response, stagedFiles, unstagedFiles)
}

// parseXMLResponse parses XML-formatted AI responses
func parseXMLResponse(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
    var groups []SuggestionGroup

    // Extract XML content from the response (it might be surrounded by other text)
    xmlPattern := regexp.MustCompile(`<suggestions>(?s).*?</suggestions>`)
    xmlMatch := xmlPattern.FindString(response)
    if xmlMatch == "" {
        return groups // No XML found
    }

    // Parse the XML
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
        Commit  Commit  `xml:"commit"`
    }

    type Suggestions struct {
        Groups []Group `xml:"group"`
    }

    var suggestions Suggestions
    err := xml.Unmarshal([]byte(xmlMatch), &suggestions)
    if err != nil {
        return groups // XML parsing failed
    }

    // Convert to SuggestionGroup objects
    for _, xmlGroup := range suggestions.Groups {
        var files []string
        for _, file := range xmlGroup.Files.File {
            files = append(files, strings.TrimSpace(file.Value))
        }

        // Validate files against the repository
        validatedFiles := make([]string, 0, len(files))
        for _, file := range files {
            if contains(stagedFiles, file) || contains(unstagedFiles, file) {
                validatedFiles = append(validatedFiles, file)
            }
        }

        // Only create a group if we have valid files
        if len(validatedFiles) > 0 {
            // Determine if files should be staged
            shouldStage := false
            for _, file := range validatedFiles {
                for _, unstagedFile := range unstagedFiles {
                    if file == unstagedFile {
                        shouldStage = true
                        break
                    }
                }
                if shouldStage {
                    break
                }
            }

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
            } else if xmlGroup.Message != "" {
                // Fallback to the legacy message field
                message = xmlGroup.Message
            }

            // Create the suggestion group
            group := SuggestionGroup{
                Description: xmlGroup.Description,
                Files:       validatedFiles,
                Message:     message,
                ShouldStage: shouldStage,
            }

            groups = append(groups, group)
        }
    }

    return groups
}

// parseRegexResponse parses AI responses using regex (legacy format)
func parseRegexResponse(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
    var groups []SuggestionGroup

    // Pattern for finding groupings - more comprehensive
    groupPattern := regexp.MustCompile(`(?i)(?:Group|Grouping|Suggestion|Commit)\s*(?:group)?\s*\d+:?\s*([^\n]+)`)
    filePattern := regexp.MustCompile(`(?i)(?:Files?|Include|Changes):\s*([^\n]+(?:\n\s*[-*]\s*[^\n]+)*)`)
    messagePattern := regexp.MustCompile(`(?i)(?:Commit message|Message|Description|Commit):\s*([^\n]+)`)
    impactPattern := regexp.MustCompile(`(?i)(?:Impact|Version impact|Semver|Version):\s*([^\n]+)`)

    // Find all groupings
    groupMatches := groupPattern.FindAllStringSubmatchIndex(response, -1)

    // Process standard groupings
    for i, groupMatch := range groupMatches {
        if len(groupMatch) < 4 {
            continue
        }

        // Extract the group description
        description := response[groupMatch[2]:groupMatch[3]]

        // Determine the end of this group (either the start of the next group or the end of the response)
        groupEnd := len(response)
        if i+1 < len(groupMatches) {
            groupEnd = groupMatches[i+1][0]
        }

        // Extract the group content
        groupContent := response[groupMatch[0]:groupEnd]

        // Extract files
        var files []string
        fileMatches := filePattern.FindStringSubmatch(groupContent)
        if len(fileMatches) > 1 {
            fileList := fileMatches[1]

            // Check if the file list contains bullet points
            if strings.Contains(fileList, "-") || strings.Contains(fileList, "*") {
                // Extract files from bullet points
                bulletPattern := regexp.MustCompile(`[-*]\s*([^\n,;]+)`)
                bulletMatches := bulletPattern.FindAllStringSubmatch(fileList, -1)
                for _, match := range bulletMatches {
                    if len(match) > 1 {
                        file := strings.TrimSpace(match[1])
                        file = regexp.MustCompile("`|'|\"").ReplaceAllString(file, "")
                        if file != "" {
                            files = append(files, file)
                        }
                    }
                }
            } else {
                // Split by commas, semicolons, or newlines
                for _, file := range regexp.MustCompile(`[,;\n]+`).Split(fileList, -1) {
                    file := strings.TrimSpace(file)
                    file = regexp.MustCompile("`|'|\"").ReplaceAllString(file, "")
                    if file != "" {
                        files = append(files, file)
                    }
                }
            }
        }

        // Extract commit message
        message := ""
        messageMatches := messagePattern.FindStringSubmatch(groupContent)
        if len(messageMatches) > 1 {
            message = strings.TrimSpace(messageMatches[1])
            message = regexp.MustCompile("`|'|\"").ReplaceAllString(message, "")
        }

        // Extract impact (optional now)
        impact := ""
        impactMatches := impactPattern.FindStringSubmatch(groupContent)
        if len(impactMatches) > 1 {
            impact = strings.TrimSpace(impactMatches[1])
            impact = regexp.MustCompile("`|'|\"").ReplaceAllString(impact, "")
        }

        // Validate files against the repository
        validatedFiles := make([]string, 0, len(files))
        for _, file := range files {
            // Check if the file exists in either staged or unstaged files
            if contains(stagedFiles, file) || contains(unstagedFiles, file) {
                validatedFiles = append(validatedFiles, file)
            }
        }

        // Only create a group if we have valid files
        if len(validatedFiles) > 0 {
            // Determine if files should be staged
            shouldStage := false
            for _, file := range validatedFiles {
                for _, unstagedFile := range unstagedFiles {
                    if file == unstagedFile {
                        shouldStage = true
                        break
                    }
                }
                if shouldStage {
                    break
                }
            }

            // Create the suggestion group
            group := SuggestionGroup{
                Description: description,
                Files:       validatedFiles,
                Message:     message,
                Impact:      impact,
                ShouldStage: shouldStage,
            }

            groups = append(groups, group)
        }
    }
    
    // If no groups were found, try fallback parsing
    if len(groups) == 0 {
        groups = fallbackParsing(response, stagedFiles, unstagedFiles)
    }

    return groups
}

// fallbackParsing tries to extract file paths and commit messages from unstructured text
func fallbackParsing(response string, stagedFiles, unstagedFiles []string) []SuggestionGroup {
    var groups []SuggestionGroup
    
    // Look for file paths with extensions
    fileExtPattern := regexp.MustCompile(`\b[\w\-./]+\.(go|js|py|java|rb|php|ts|jsx|tsx|html|css|md|json|yaml|yml|toml|xml|sql|sh|bat|c|cpp|h|hpp)\b`)
    extMatches := fileExtPattern.FindAllString(response, -1)

    var files []string
    for _, match := range extMatches {
        match = strings.TrimSpace(match)
        match = regexp.MustCompile("`|'|\"").ReplaceAllString(match, "")
        if match != "" && !contains(files, match) {
            files = append(files, match)
        }
    }

    // Look for commit message
    message := ""
    conventionalPattern := regexp.MustCompile(`(?i)(feat|fix|docs|style|refactor|test|chore)(?:\(([^)]+)\))?:\s*([^\n]+)`)
    conventionalMatches := conventionalPattern.FindStringSubmatch(response)
    if len(conventionalMatches) > 3 {
        scope := ""
        if conventionalMatches[2] != "" {
            scope = "(" + conventionalMatches[2] + ")"
        }
        message = conventionalMatches[1] + scope + ": " + strings.TrimSpace(conventionalMatches[3])
    }

    if len(files) > 0 || message != "" {
        shouldStage := false
        for _, file := range files {
            for _, unstagedFile := range unstagedFiles {
                if file == unstagedFile {
                    shouldStage = true
                    break
                }
            }
            if shouldStage {
                break
            }
        }

        groups = append(groups, SuggestionGroup{
            Description: "Suggested changes",
            Files:       files,
            Message:     message,
            ShouldStage: shouldStage,
        })
    }
    
    return groups
}
// contains checks if a string is in a slice
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
	return strings.Split(strings.TrimSpace(s), "\n")
}

// ExecuteCommand executes a shell command and returns its output
func ExecuteCommand(cmd string) (string, error) {
	// Split the command into parts
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil
	}

	// Create a new command
	command := parts[0]
	args := parts[1:]

	// Execute the command
	output, err := exec.Command(command, args...).Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
