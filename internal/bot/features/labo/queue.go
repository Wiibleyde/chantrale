package labo

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type LaboEntry struct {
	ChannelID string
	MessageID string
	UserID    string
	StartTime time.Time
	Name      string
	Type      string
	Result    string
	Time      int
}

type LaboQueue struct {
	mu      sync.Mutex
	entries []*LaboEntry
	ticker  *time.Ticker
	session *discordgo.Session
}

var Queue = &LaboQueue{}

func (q *LaboQueue) SetSession(s *discordgo.Session) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.session = s
}

func (q *LaboQueue) Add(entry *LaboEntry) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.entries = append(q.entries, entry)

	if q.ticker == nil {
		q.ticker = time.NewTicker(10 * time.Second)
		go q.run()
	}
}

func (q *LaboQueue) run() {
	for range q.ticker.C {
		q.mu.Lock()
		now := time.Now()
		var remaining []*LaboEntry
		var toNotify []*LaboEntry

		for _, e := range q.entries {
			elapsed := now.Sub(e.StartTime)
			if elapsed >= time.Duration(e.Time)*time.Minute {
				toNotify = append(toNotify, e)
			} else {
				remaining = append(remaining, e)
			}
		}
		q.entries = remaining

		if len(q.entries) == 0 && q.ticker != nil {
			q.ticker.Stop()
			q.ticker = nil
		}

		session := q.session
		q.mu.Unlock()

		for _, e := range toNotify {
			if session != nil {
				notifyCompletion(session, e)
			}
		}

		if q.ticker == nil {
			return
		}
	}
}

func (q *LaboQueue) CancelByMessageID(messageID string) (bool, *LaboEntry) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for idx, e := range q.entries {
		if e.MessageID == messageID {
			q.entries = append(q.entries[:idx], q.entries[idx+1:]...)
			if len(q.entries) == 0 && q.ticker != nil {
				q.ticker.Stop()
				q.ticker = nil
			}
			return true, e
		}
	}
	return false, nil
}

func notifyCompletion(s *discordgo.Session, entry *LaboEntry) {
	embed := BuildLaboResultEmbed(entry)
	components := []discordgo.MessageComponent{}

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         entry.MessageID,
		Channel:    entry.ChannelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("Error editing labo message: %v", err)
	}

	ping, err := s.ChannelMessageSend(entry.ChannelID, fmt.Sprintf("<@%s> Votre analyse est terminée !", entry.UserID))
	if err != nil {
		log.Printf("Error sending ping: %v", err)
		return
	}

	time.AfterFunc(60*time.Second, func() {
		if err := s.ChannelMessageDelete(entry.ChannelID, ping.ID); err != nil {
			log.Printf("Error deleting ping: %v", err)
		}
	})
}
