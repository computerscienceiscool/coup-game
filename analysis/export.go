package analysis

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/computerscienceiscool/coup-game/simulation"
)

// GenerateCSVs creates all required CSV reports
func GenerateCSVs(stats *StatisticsResult, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each required CSV file
	if err := generateCharacterStats(stats, outputDir); err != nil {
		return err
	}

	if err := generateGameLogs(stats, outputDir); err != nil {
		return err
	}

	if err := generateActionLogs(stats, outputDir); err != nil {
		return err
	}

	if err := generatePlayerCountAnalysis(stats, outputDir); err != nil {
		return err
	}

	return nil
}

// generateCharacterStats creates the character_stats.csv file
func generateCharacterStats(stats *StatisticsResult, outputDir string) error {
	filePath := filepath.Join(outputDir, "character_stats.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create character stats file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Character",
		"WinRate",
		"ActionSuccessRate",
		"SurvivalTime",
		"BluffSuccessRate",
		"ChallengeSuccessRate",
		"BlockSuccessRate",
		"TimesUsed",
		"TimesWon",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write character data
	for _, char := range stats.RankedCharacters {
		charStats := stats.CharacterStats[char.Name]
		row := []string{
			charStats.Name,
			formatPercent(charStats.WinRate),
			formatPercent(charStats.ActionSuccessRate),
			fmt.Sprintf("%.2f", charStats.SurvivalTime),
			formatPercent(charStats.BluffSuccessRate),
			formatPercent(charStats.ChallengeSuccessRate),
			formatPercent(charStats.BlockSuccessRate),
			strconv.Itoa(charStats.TimesUsed),
			strconv.Itoa(charStats.TimesWon),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing character row: %w", err)
		}
	}

	return nil
}

// generateGameLogs creates the game_logs.csv file
func generateGameLogs(stats *StatisticsResult, outputDir string) error {
	filePath := filepath.Join(outputDir, "game_logs.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create game logs file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"GameID",
		"PlayerCount",
		"Winner",
		"WinnerCharacters",
		"TotalTurns",
		"Date",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Get the simulation results from stats
	results := simulationResults(stats)
	if results == nil {
		return fmt.Errorf("simulation results not available")
	}

	// Write game data
	for _, result := range results.Results {
		row := []string{
			strconv.Itoa(result.ID),
			strconv.Itoa(result.PlayerCount),
			strconv.Itoa(result.WinnerID),
			formatCharacters(result.WinnerCharacters),
			strconv.Itoa(result.TotalTurns),
			result.EndTime.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing game row: %w", err)
		}
	}

	return nil
}

// generateActionLogs creates the action_logs.csv file
func generateActionLogs(stats *StatisticsResult, outputDir string) error {
	filePath := filepath.Join(outputDir, "action_logs.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create action logs file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"GameID",
		"Turn",
		"Player",
		"Action",
		"Target",
		"Success",
		"Challenged",
		"Blocked",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Get the simulation results from stats
	results := simulationResults(stats)
	if results == nil {
		return fmt.Errorf("simulation results not available")
	}

	// Write action data (limit to first 10k actions to avoid huge files)
	actionCount := 0
	maxActions := 10000

	for _, game := range results.Results {
		for _, action := range game.Actions {
			// Skip if we've hit the max
			if actionCount >= maxActions {
				break
			}

			target := "-"
			if action.Target >= 0 {
				target = strconv.Itoa(action.Target)
			}

			blocker := "No"
			if action.Blocker >= 0 {
				blocker = strconv.Itoa(action.Blocker)
			}

			row := []string{
				strconv.Itoa(game.ID),
				strconv.Itoa(action.Turn),
				strconv.Itoa(action.PlayerID),
				action.Action,
				target,
				formatBool(action.Success),
				formatBool(action.Challenged),
				blocker,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("error writing action row: %w", err)
			}

			actionCount++
		}
	}

	return nil
}

// generatePlayerCountAnalysis creates the player_count_analysis.csv file
func generatePlayerCountAnalysis(stats *StatisticsResult, outputDir string) error {
	filePath := filepath.Join(outputDir, "player_count_analysis.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create player count analysis file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"PlayerCount",
		"Character",
		"WinRate",
		"AvgSurvivalTime",
		"AvgGameLength",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write player count data
	for count, pcStats := range stats.PlayerCountStats {
		for char, winRate := range pcStats.CharacterWinRates {
			// Get character stats
			charStats := stats.CharacterStats[char]

			row := []string{
				strconv.Itoa(count),
				char,
				formatPercent(winRate),
				fmt.Sprintf("%.2f", charStats.SurvivalTime),
				fmt.Sprintf("%.2f", pcStats.AverageGameLength),
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("error writing player count row: %w", err)
			}
		}
	}

	return nil
}

// Helper functions for formatting

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

func formatBool(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func formatCharacters(chars []string) string {
	result := ""
	for i, char := range chars {
		if i > 0 {
			result += ","
		}
		result += char
	}
	return result
}

// simulationResults retrieves the simulation.SimulationResults from stats
// This is a hack since we don't have direct access to the original results
func simulationResults(stats *StatisticsResult) *simulation.SimulationResults {
	// In a real implementation, we would have proper access to the results
	// For now, return nil and handle the error
	return nil
}
