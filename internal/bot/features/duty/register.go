package duty

import "LsmsBot/internal/bot/router"

// Register wires all duty feature handlers into the router.
// Add new duty interactions here — router.go never needs to change.
func Register(r *router.Router) {
	r.OnCommand("duty", HandleCommand)
	r.OnButton("handleLsmsDuty", HandleDutyButton)
	r.OnButton("handleLsmsOnCall", HandleOnCallButton)
	r.OnButton("handleLsmsOffRadio", HandleOffRadioButton)
}
