package game

type State interface {
	Update(g *Game, p *Player)
}
