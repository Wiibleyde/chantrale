package radio

import "LsmsBot/internal/bot/router"

// Register wires all radio feature handlers into the router.
func Register(r *router.Router) {
	r.OnCommand("radio", HandleCommand)
	r.OnButton("lsmsRadioAdd", HandleRadioAdd)
	r.OnButton("lsmsRadioRemove", HandleRadioRemove)
	r.OnButtonPrefix("lsmsRadioEdit--", HandleRadioEdit)
	r.OnButtonPrefix("lsmsRadioRemoveSelect--", HandleRadioRemoveSelect)
	r.OnModal("lsmsRadioAddModal--", HandleRadioAddModal)
	r.OnModal("lsmsRadioEditModal--", HandleRadioEditModal)
}
