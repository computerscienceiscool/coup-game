package analysis

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/computerscienceiscool/coup-game/simulation"
)

// GenerateCSVs creates all required CSV reports
func GenerateCSVs(stats *StatisticsResult, results *simulation.SimulationResults, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each required CSV file
	if err := generateCharacterStats(stats, outputDir); err != nil {
		return err
	}

	if err := generateGameLogs(results, outputDir); err != nil {
		return err
	}

	if err := generateActionLogs(results, outputDir); err != nil {
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

	// Write header. Metric definitions live in docs/specification.md.
	header := []string{
		"Character",
		"DealtWinRate",      // P(win | dealt the character at game start)
		"FinalHandWinRate",  // share of games whose winner ended holding it
		"ActionSuccessRate", // signature action successes/attempts
		"BlockSuccessRate",  // blocks claiming it that stopped the action
		"BluffRate",         // share of its claims made without the card
		"BluffSuccessRate",  // bluffed claims that went unchallenged
		"ChallengedRate",    // share of its claims that were challenged
		"AvgTurnsSurvived",  // average turns survived by players dealt it
		"TimesDealt",
		"WinsWhenDealt",
		"PowerScore",        // composite: 0.6*win + 0.15*action + 0.1*block + 0.1*bluff + 0.05*survival
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write character data
	for _, char := range stats.RankedCharacters {
		charStats := stats.CharacterStats[char.Name]
		row := []string{
			charStats.Name,
			formatPercent(charStats.DealtWinRate),
			formatPercent(charStats.FinalHandWinRate),
			formatPercent(charStats.ActionSuccessRate),
			formatPercent(charStats.BlockSuccessRate),
			formatPercent(charStats.BluffRate),
			formatPercent(charStats.BluffSuccessRate),
			formatPercent(charStats.ChallengedRate),
			fmt.Sprintf("%.2f", charStats.SurvivalTime),
			strconv.Itoa(charStats.TimesDealt),
			strconv.Itoa(charStats.WinsWhenDealt),
			fmt.Sprintf("%.4f", char.PowerScore),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing character row: %w", err)
		}
	}

	return nil
}

// generateGameLogs creates the game_logs.csv file
func generateGameLogs(results *simulation.SimulationResults, outputDir string) error {
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
func generateActionLogs(results *simulation.SimulationResults, outputDir string) error {
	filePath := filepath.Join(outputDir, "action_logs.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create action logs file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header. Blocker records every block attempt (including blocks
	// later defeated by a challenge); BlockSucceeded says whether the block
	// stopped the action. ActorHadCard/BlockerHadCard are ground truth for
	// bluff analysis and are never shown to AI players.
	header := []string{
		"GameID",
		"Turn",
		"Player",
		"Action",
		"Target",
		"Success",
		"Challenged",
		"ActorHadCard",
		"Blocker",
		"BlockingCharacter",
		"BlockChallenged",
		"BlockSucceeded",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write all action data
	for _, game := range results.Results {
		for _, action := range game.Actions {
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
				formatBool(action.ActorHadCard),
				blocker,
				action.BlockingCharacter,
				formatBool(action.BlockChallenged),
				formatBool(action.BlockSucceeded),
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("error writing action row: %w", err)
			}
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

	// Write header. DealtWinRate is per player count; AvgGameLength is the
	// measured average number of actions per game for that player count.
	header := []string{
		"PlayerCount",
		"Character",
		"DealtWinRate",
		"AvgGameLength",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write player count data (sorted for deterministic output)
	playerCounts := make([]int, 0, len(stats.PlayerCountStats))
	for count := range stats.PlayerCountStats {
		playerCounts = append(playerCounts, count)
	}
	sort.Ints(playerCounts)

	for _, count := range playerCounts {
		pcStats := stats.PlayerCountStats[count]

		charNames := make([]string, 0, len(pcStats.CharacterWinRates))
		for char := range pcStats.CharacterWinRates {
			charNames = append(charNames, char)
		}
		sort.Strings(charNames)

		for _, char := range charNames {
			row := []string{
				strconv.Itoa(count),
				char,
				formatPercent(pcStats.CharacterWinRates[char]),
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

