package legacy

/*
#include "GAME1.h"
*/
import "C"

// SetOpenworldNewGame toggles the openworld start-map remap for new games
// (set by the Open World main menu entry, reset whenever the menu is shown).
func SetOpenworldNewGame(v bool) {
	if v {
		C.nox_openworld_newgame = 1
	} else {
		C.nox_openworld_newgame = 0
	}
}
