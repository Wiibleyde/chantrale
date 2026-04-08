package doctor

import "LsmsBot/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("doctor", HandleCommand)
}
