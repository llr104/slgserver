package model

var ArmyIsInView func(rid, x, y int) bool
var GetUnionId func(rid int) int
var GetUnionName func(unionId int) string
var GetRoleNickName func(rid int) string
var GetParentId func(rid int) int
var GetMainMembers func(unionId int) []int