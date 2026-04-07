package router

import (
"strings"

"github.com/disgoorg/disgo/bot"
"github.com/disgoorg/disgo/events"
)

type CommandHandler func(e *events.ApplicationCommandInteractionCreate)
type ComponentHandler func(e *events.ComponentInteractionCreate)
type ModalHandler func(e *events.ModalSubmitInteractionCreate)

type prefixEntry[H any] struct {
prefix  string
handler H
}

type Router struct {
commands       map[string]CommandHandler
buttons        map[string]ComponentHandler
buttonPrefixes []prefixEntry[ComponentHandler]
modalPrefixes  []prefixEntry[ModalHandler]
}

func New() *Router {
return &Router{
commands: make(map[string]CommandHandler),
buttons:  make(map[string]ComponentHandler),
}
}

func (r *Router) OnCommand(name string, h CommandHandler) { r.commands[name] = h }
func (r *Router) OnButton(id string, h ComponentHandler)  { r.buttons[id] = h }
func (r *Router) OnButtonPrefix(prefix string, h ComponentHandler) {
r.buttonPrefixes = append(r.buttonPrefixes, prefixEntry[ComponentHandler]{prefix, h})
}
func (r *Router) OnModal(prefix string, h ModalHandler) {
r.modalPrefixes = append(r.modalPrefixes, prefixEntry[ModalHandler]{prefix, h})
}

func (r *Router) Attach(client *bot.Client) {
client.AddEventListeners(
bot.NewListenerFunc(func(e *events.ApplicationCommandInteractionCreate) {
name := e.SlashCommandInteractionData().CommandName()
if h, ok := r.commands[name]; ok {
h(e)
}
}),
bot.NewListenerFunc(func(e *events.ComponentInteractionCreate) {
id := e.Data.CustomID()
if h, ok := r.buttons[id]; ok {
h(e)
return
}
for _, pe := range r.buttonPrefixes {
if strings.HasPrefix(id, pe.prefix) {
pe.handler(e)
return
}
}
}),
bot.NewListenerFunc(func(e *events.ModalSubmitInteractionCreate) {
id := e.Data.CustomID
for _, pe := range r.modalPrefixes {
if strings.HasPrefix(id, pe.prefix) {
pe.handler(e)
return
}
}
}),
)
}
