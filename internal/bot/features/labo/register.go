package labo

import "LsmsBot/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("labo", HandleCommand)
	r.OnButton("laboCancelButton", HandleCancelButton)
}
