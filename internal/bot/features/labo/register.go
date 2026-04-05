package labo

import "LsmsBot/internal/bot/router"

// Register wires all labo feature handlers into the router.
func Register(r *router.Router) {
	r.OnCommand("labo", HandleCommand)
	r.OnButton("laboCancelButton", HandleCancelButton)
}
