package game

import (
	"fmt"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"time"
)

var (
	RotationBase = vec.Right
)

const (
	NormalSpeed = 2.5

	FullChargeSpeed    = 0.2 * NormalSpeed
	FullChargeDuration = time.Second

	DashSpeed    = 2 * NormalSpeed
	DashDuration = 250 * time.Millisecond
	DashCooldown = time.Second
)

type Player struct {
	Pos           vec.Vec2
	Velocity      vec.Vec2
	MovementState State
	PlayerState   State
	Rotation      float64
}

func NewPlayer(x, y float64) *Player {
	return &Player{
		Pos:      vec.Vec2{X: x, Y: y},
		Velocity: vec.Vec2{},
		MovementState: NormalMovementState{
			lastDashTime: time.Now().Add(-DashCooldown),
		},
		PlayerState: NormalPlayerState{},
	}
}

func (p *Player) Update(g *Game) {
	moveVector := vec.Vec2{}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		moveVector.Add(-1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		moveVector.Add(1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		moveVector.Add(0, -1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		moveVector.Add(0, 1)
	}

	mx, my := ebiten.CursorPosition()
	// TODO Radius 32
	cp := vec.NewVec2(p.Pos.X+16, p.Pos.Y+16)
	mVec := vec.NewIVec2(mx, my)
	if mVec.Y > cp.Y {
		mVec.SubVec(cp)
		p.Rotation = math.Acos(vec.Right.Dot(mVec) / (vec.Right.Len() * mVec.Len()))
	} else {
		mVec.SubVec(cp)
		p.Rotation = math.Pi + math.Acos(-vec.Right.Dot(mVec)/(vec.Right.Len()*mVec.Len()))
	}

	fmt.Println("Player rotation:", p.Rotation)

	moveVector.Normalize()
	p.Velocity = moveVector

	p.MovementState.Update(g, p)
	p.PlayerState.Update(g, p)

	p.Pos.AddVec(p.Velocity)

	// Wall collisions
	if p.Pos.X <= 0 {
		p.Pos.X = 0
	}
	if p.Pos.X+32 >= constants.ScreenWidth {
		p.Pos.X = constants.ScreenWidth - 32
	}
	if p.Pos.Y <= 0 {
		p.Pos.Y = 0
	}
	if p.Pos.Y+32 >= constants.ScreenHeight {
		p.Pos.Y = constants.ScreenHeight - 32
	}

	// TODO ball wall conditions (now one can throw ball outside of the screen) ;/
}
