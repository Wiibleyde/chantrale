// Package router provides a self-registering interaction router for discordgo.
// Features call Register(r) once; the router handles all dispatch internally.
// Adding a new feature only requires one line in bot.go — this file never changes.
package router

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Handler is the standard signature for all interaction handlers.
type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)

type prefixEntry struct {
	prefix  string
	handler Handler
}

// Router dispatches Discord interactions to registered handlers.
type Router struct {
	commands       map[string]Handler // exact command name
	buttons        map[string]Handler // exact button/select customID
	buttonPrefixes []prefixEntry      // prefix-matched component customIDs
	modalPrefixes  []prefixEntry      // prefix-matched modal customIDs
}

// New creates a ready-to-use Router.
func New() *Router {
	return &Router{
		commands: make(map[string]Handler),
		buttons:  make(map[string]Handler),
	}
}

// OnCommand registers a handler for a slash command by exact name.
func (r *Router) OnCommand(name string, h Handler) {
	r.commands[name] = h
}

// OnButton registers a handler for an exact component customID.
func (r *Router) OnButton(id string, h Handler) {
	r.buttons[id] = h
}

// OnButtonPrefix registers a handler for any component customID that starts with prefix.
func (r *Router) OnButtonPrefix(prefix string, h Handler) {
	r.buttonPrefixes = append(r.buttonPrefixes, prefixEntry{prefix, h})
}

// OnModal registers a handler for any modal customID that starts with prefix.
func (r *Router) OnModal(prefix string, h Handler) {
	r.modalPrefixes = append(r.modalPrefixes, prefixEntry{prefix, h})
}

// Attach wires the router's dispatch logic into the Discord session.
// Call this once after all features have registered.
func (r *Router) Attach(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {

		case discordgo.InteractionApplicationCommand:
			name := i.ApplicationCommandData().Name
			if h, ok := r.commands[name]; ok {
				h(s, i)
			}

		case discordgo.InteractionMessageComponent:
			id := i.MessageComponentData().CustomID
			if h, ok := r.buttons[id]; ok {
				h(s, i)
				return
			}
			for _, pe := range r.buttonPrefixes {
				if strings.HasPrefix(id, pe.prefix) {
					pe.handler(s, i)
					return
				}
			}

		case discordgo.InteractionModalSubmit:
			id := i.ModalSubmitData().CustomID
			for _, pe := range r.modalPrefixes {
				if strings.HasPrefix(id, pe.prefix) {
					pe.handler(s, i)
					return
				}
			}
		}
	})
}
