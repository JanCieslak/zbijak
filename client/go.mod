module github.com/JanCieslak/Zbijak/client

go 1.18

require (
	github.com/JanCieslak/zbijak/common v0.0.0-20220219130121-a4850142f605
	github.com/hajimehoshi/ebiten/v2 v2.2.4
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
)

replace github.com/JanCieslak/zbijak/common => ../common

require (
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20210727001814-0db043d8d5be // indirect
	github.com/jezek/xgb v0.0.0-20210312150743-0e0f116e1240 // indirect
	golang.org/x/exp v0.0.0-20190731235908-ec7cb31e5a56 // indirect
	golang.org/x/mobile v0.0.0-20210902104108-5d9a33257ab5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
)
