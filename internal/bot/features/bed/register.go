package bed

import "LsmsBot/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("beds", HandleCommand)
	r.OnButtonPrefix("lsmsBed--", HandleRemoveBed)
}
