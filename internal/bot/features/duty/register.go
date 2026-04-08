package duty

import "LsmsBot/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("duty", HandleCommand)
	r.OnButton("handleLsmsDuty", HandleDutyButton)
	r.OnButton("handleLsmsOnCall", HandleOnCallButton)
	r.OnButton("handleLsmsOffRadio", HandleOffRadioButton)
}
