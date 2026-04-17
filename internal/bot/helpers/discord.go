package helpers

import (
	"strings"

	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// DeleteMessageByIDs parses channel/message ID strings and deletes the message.
func DeleteMessageByIDs(client *bot.Client, channelID, messageID string) {
	chanID, err := snowflake.Parse(channelID)
	if err != nil {
		return
	}
	msgID, err := snowflake.Parse(messageID)
	if err != nil {
		return
	}
	if err := client.Rest.DeleteMessage(chanID, msgID); err != nil {
		logger.Error("Error deleting message", "error", err, "channelID", channelID, "messageID", messageID)
	}
}

// SuffixFromCustomID extracts the part after the first "--" separator.
func SuffixFromCustomID(customID string) string {
	parts := strings.SplitN(customID, "--", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// BuildActionRows groups interactive components into rows of at most perRow each.
func BuildActionRows(buttons []discord.InteractiveComponent, perRow int) []discord.LayoutComponent {
	if len(buttons) == 0 {
		return []discord.LayoutComponent{}
	}
	var rows []discord.LayoutComponent
	for i := 0; i < len(buttons); i += perRow {
		end := i + perRow
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, discord.ActionRowComponent{Components: buttons[i:end]})
	}
	return rows
}
