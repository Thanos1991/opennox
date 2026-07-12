package opennox

import (
	"strings"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

// script.Factions implementation. The open-world layer uses these to make the
// host player's own class faction friendly (same team) while leaving other
// factions hostile (different team). See Server.IsEnemyTo: units are enemies
// unless they share a team.

func (s noxScriptImpl) HostPlayerClass() string {
	pl := s.s.Players.ByInd(server.HostPlayerIndex)
	if pl == nil {
		return ""
	}
	switch pl.PlayerClass() {
	case player.Warrior:
		return "warrior"
	case player.Wizard:
		return "wizard"
	case player.Conjurer:
		return "conjurer"
	}
	return ""
}

// teamByName returns the persistent team with the given name, creating it if
// needed. Teams reset on map load, so lookup-then-create keeps script calls
// idempotent across zones.
func (s noxScriptImpl) teamByName(name string) *server.Team {
	if name == "" {
		return nil
	}
	srv := s.s
	for t := srv.Teams.First(); t != nil; t = srv.Teams.Next(t) {
		if strings.EqualFold(t.Name(), name) {
			return t
		}
	}
	t := srv.Teams.Create(0) // 0 => allocate a free ID
	if t == nil {
		return nil
	}
	t.SetNameAnd68(name, 0)
	return t
}

func (s noxScriptImpl) SetHostPlayerTeam(name string) {
	tm := s.teamByName(name)
	if tm == nil {
		return
	}
	pl := s.s.Players.ByInd(server.HostPlayerIndex)
	if pl == nil || pl.PlayerUnit == nil {
		return
	}
	asObjectS(pl.PlayerUnit).SetTeam(tm)
}

func (s noxScriptImpl) SetAllUnitsTeam(name string) {
	tm := s.teamByName(name)
	if tm == nil {
		return
	}
	for obj := s.s.Objs.First(); obj != nil; obj = obj.Next() {
		// monsters/NPCs only; never re-team players
		if !obj.Class().HasAny(object.ClassMonster) {
			continue
		}
		asObjectS(obj).SetTeam(tm)
	}
}
