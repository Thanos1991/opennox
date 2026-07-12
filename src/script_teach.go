package opennox

import (
	"github.com/opennox/libs/spell"
	ns4 "github.com/opennox/noxscript/ns/v4"
	nsp "github.com/opennox/noxscript/ns/v4/spell"
)

// script.SpellTeacher implementation. Spawns a spell-reward book (the engine's
// own teaching item) that permanently grants the spell to a compatible-class
// player on pickup. Used by the open-world academy classrooms.
func (s noxScriptImpl) GiveSpellBook(x, y float32, spellName string) bool {
	if !spell.ParseID(spellName).Valid() {
		return false
	}
	obj := s.s.noxScriptP().NewSpellBook(ns4.Ptf(x, y), nsp.Spell(spellName))
	return obj != nil
}
