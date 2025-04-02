package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// HistoryContext represents historical analysis of file changes
type HistoryContext struct {
	Patterns     map[string]*ChangePattern // Path -> Pattern mapping
	LastAnalysis time.Time
	StartTime    time.Time
	EndTime      time.Time
}

// ChangePattern tracks how a file or component changes over time
type ChangePattern struct {
	Path          string
	ChangeCount   int
	LastModified  time.Time
	FirstModified time.Time
	Contributors  map[string]int     // Author -> Commit count
	RelatedFiles  map[string]float64 // Related file -> Correlation score
	ChangeTypes   map[string]int     // Type of change -> Count
	Complexity    []ComplexityMetric // Historical complexity metrics
	ImpactScore   float64            // Calculated impact score
	Stability     float64            // Change frequency stability (0-1)
}

// ComplexityMetric tracks complexity over time
type ComplexityMetric struct {
	Timestamp  time.Time
	Score      int
	Churn      int // Lines changed
	Author     string
	CommitHash string
}

// HistoryAnalyzer handles git history analysis
type HistoryAnalyzer struct {
	repo        *git.Repository
	maxHistory  time.Duration
	minPatterns int
}

// NewHistoryAnalyzer creates a new history analyzer
func NewHistoryAnalyzer(repoPath string) (*HistoryAnalyzer, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &HistoryAnalyzer{
		repo:        repo,
		maxHistory:  90 * 24 * time.Hour, // 90 days default
		minPatterns: 3,                   // Minimum patterns to establish correlation
	}, nil
}

// AnalyzeHistory performs historical analysis of the repository
func (h *HistoryAnalyzer) AnalyzeHistory(ctx context.Context) (*HistoryContext, error) {
	head, err := h.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Initialize history context
	histCtx := &HistoryContext{
		Patterns:     make(map[string]*ChangePattern),
		LastAnalysis: time.Now(),
		StartTime:    time.Now().Add(-h.maxHistory),
		EndTime:      time.Now(),
	}

	// Get commit iterator
	cIter, err := h.repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit iterator: %w", err)
	}
	defer cIter.Close()

	// Process each commit
	err = cIter.ForEach(func(c *object.Commit) error {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip if commit is too old
		if c.Committer.When.Before(histCtx.StartTime) {
			return nil
		}

		// Get commit stats
		stats, err := c.Stats()
		if err != nil {
			return fmt.Errorf("failed to get commit stats: %w", err)
		}

		// Process each changed file
		for _, stat := range stats {
			pattern := histCtx.getOrCreatePattern(stat.Name)
			pattern.ChangeCount++
			pattern.LastModified = c.Committer.When
			if pattern.FirstModified.IsZero() {
				pattern.FirstModified = c.Committer.When
			}

			// Update contributors
			pattern.Contributors[c.Author.Email]++

			// Update change types
			changeType := h.classifyChange(stat)
			pattern.ChangeTypes[changeType]++

			// Calculate complexity metric
			complexity := ComplexityMetric{
				Timestamp:  c.Committer.When,
				Churn:      stat.Addition + stat.Deletion,
				Author:     c.Author.Email,
				CommitHash: c.Hash.String(),
			}
			pattern.Complexity = append(pattern.Complexity, complexity)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process commits: %w", err)
	}

	// Calculate correlations and impact scores
	h.calculateCorrelations(histCtx)
	h.calculateImpactScores(histCtx)

	return histCtx, nil
}

// Helper methods

func (h *HistoryContext) getOrCreatePattern(path string) *ChangePattern {
	pattern, exists := h.Patterns[path]
	if !exists {
		pattern = &ChangePattern{
			Path:         path,
			Contributors: make(map[string]int),
			RelatedFiles: make(map[string]float64),
			ChangeTypes:  make(map[string]int),
			Complexity:   make([]ComplexityMetric, 0),
		}
		h.Patterns[path] = pattern
	}
	return pattern
}

func (h *HistoryAnalyzer) classifyChange(stat object.FileStat) string {
	ratio := float64(stat.Addition) / float64(stat.Addition+stat.Deletion)
	switch {
	case stat.Addition == 0 && stat.Deletion > 0:
		return "deletion"
	case stat.Addition > 0 && stat.Deletion == 0:
		return "addition"
	case ratio > 0.7:
		return "major_addition"
	case ratio < 0.3:
		return "major_deletion"
	default:
		return "modification"
	}
}

func (h *HistoryAnalyzer) calculateCorrelations(ctx *HistoryContext) {
	// Build a timeline of changes for each file
	timeline := make(map[string][]time.Time)
	for path, pattern := range ctx.Patterns {
		for _, metric := range pattern.Complexity {
			timeline[path] = append(timeline[path], metric.Timestamp)
		}
	}

	// Calculate correlations between files
	for path1, times1 := range timeline {
		for path2, times2 := range timeline {
			if path1 == path2 {
				continue
			}
			correlation := h.calculateTimelineCorrelation(times1, times2)
			if correlation > 0.5 { // Threshold for correlation
				ctx.Patterns[path1].RelatedFiles[path2] = correlation
			}
		}
	}
}

func (h *HistoryAnalyzer) calculateImpactScores(ctx *HistoryContext) {
	for _, pattern := range ctx.Patterns {
		// Calculate impact based on:
		// 1. Change frequency
		changeFreq := float64(pattern.ChangeCount) / float64(ctx.EndTime.Sub(ctx.StartTime).Hours()/24)

		// 2. Number of contributors
		contributorScore := float64(len(pattern.Contributors)) / 5 // Normalize by expected team size

		// 3. Related file count
		relationScore := float64(len(pattern.RelatedFiles)) / float64(len(ctx.Patterns))

		// 4. Complexity trend
		complexityTrend := h.calculateComplexityTrend(pattern.Complexity)

		// Combine scores with weights
		pattern.ImpactScore = (changeFreq * 0.3) +
			(contributorScore * 0.2) +
			(relationScore * 0.2) +
			(complexityTrend * 0.3)

		// Calculate stability (inverse of change frequency, normalized)
		pattern.Stability = 1 / (1 + changeFreq)
	}
}

func (h *HistoryAnalyzer) calculateTimelineCorrelation(times1, times2 []time.Time) float64 {
	// Simple temporal correlation based on change proximity
	matches := 0
	threshold := 24 * time.Hour // Consider changes related if within 24 hours

	for _, t1 := range times1 {
		for _, t2 := range times2 {
			if t1.Sub(t2).Abs() < threshold {
				matches++
			}
		}
	}

	return float64(matches) / float64(len(times1)+len(times2))
}

func (h *HistoryAnalyzer) calculateComplexityTrend(metrics []ComplexityMetric) float64 {
	if len(metrics) < 2 {
		return 0
	}

	// Calculate the trend of complexity over time
	var trend float64
	for i := 1; i < len(metrics); i++ {
		if metrics[i].Churn > metrics[i-1].Churn {
			trend++
		} else {
			trend--
		}
	}

	return trend / float64(len(metrics)-1)
}
