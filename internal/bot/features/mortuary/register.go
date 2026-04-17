package mortuary

import "LsmsBot/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("mortuary", HandleCommand)
	r.OnButtonPrefix("lsmsMortuary--", HandleRemoveLocker)
}
