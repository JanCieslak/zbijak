module github.com/JanCieslak/zbijak/server

go 1.16

replace (
	github.com/JanCieslak/zbijak/common => ../common
)

require (
	github.com/JanCieslak/zbijak/common v0.0.0-20220218162710-1577d824ae5c
	github.com/google/uuid v1.3.0
)
