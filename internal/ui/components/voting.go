package components

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

// VoteSuccessMsg is sent when a vote is successful
type VoteSuccessMsg struct {
	Message     string
	StationUUID string
}

// VoteFailedMsg is sent when a vote fails
type VoteFailedMsg struct {
	Err error
}

// ExecuteVote performs the voting logic with local persistence and API calls
func ExecuteVote(station *api.Station, votedStations *storage.VotedStations, apiClient *api.Client) tea.Cmd {
	return func() tea.Msg {
		if station == nil {
			return VoteFailedMsg{Err: fmt.Errorf("no station selected")}
		}

		// Guard against nil votedStations
		if votedStations == nil {
			return VoteFailedMsg{Err: fmt.Errorf("voting system not initialized")}
		}

		// Check if can vote (respects 10-minute API cooldown)
		if !votedStations.CanVoteAgain(station.StationUUID) {
			return VoteFailedMsg{Err: fmt.Errorf("already voted for this station (wait 10 minutes)")}
		}

		result, err := apiClient.Vote(context.Background(), station.StationUUID)

		// Check if API says we already voted (even if our local record doesn't have it)
		if err != nil || !result.OK {
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			} else {
				errMsg = result.Message
			}

			// If API says "already voted", "too often", or "VoteError", record it locally
			errMsgLower := strings.ToLower(errMsg)
			if strings.Contains(errMsgLower, "too often") ||
				strings.Contains(errMsgLower, "already voted") ||
				strings.Contains(errMsgLower, "voteerror") {
				if err := votedStations.AddVote(station.StationUUID); err != nil {
					// Failed to save vote locally - return error
					return VoteFailedMsg{Err: fmt.Errorf("failed to record vote: %w", err)}
				}
				return VoteSuccessMsg{Message: "You voted", StationUUID: station.StationUUID}
			}

			if err != nil {
				return VoteFailedMsg{Err: err}
			}
			return VoteFailedMsg{Err: fmt.Errorf("%s", errMsg)}
		}

		// Successful vote - mark as voted
		if err := votedStations.AddVote(station.StationUUID); err != nil {
			// API vote succeeded but local save failed - inform user
			return VoteFailedMsg{Err: fmt.Errorf("vote succeeded but failed to save locally: %w", err)}
		}

		return VoteSuccessMsg{Message: "Voted for " + station.TrimName(), StationUUID: station.StationUUID}
	}
}
