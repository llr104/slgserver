package model

var ArmyIsInView func(rid, x, y int) bool
var GetUnionId func(rid int) int
var GetUnionName func(unionId int) string
var GetRoleNickName func(rid int) string
var GetParentId func(rid int) int
var GetMainMembers func(unionId int) []int
var GetYield func(rid int) Yield
var GetDepotCapacity func(rid int) int
var GetCityCost func(cid int) int8
var GetMaxDurable func(cid int) int
var GetCityLv func(cid int) int8
var MapResTypeLevel func(x, y int) (bool, int8, int8)

var ServerId = 0