package doctor

import "LsmsBot/internal/bot/router"

// Register wires all doctor feature handlers into the router.
func Register(r *router.Router) {
	r.OnCommand("doctor", HandleCommand)
}
