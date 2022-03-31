package player

type State interface {
	Update(p *Player)
}
