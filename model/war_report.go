package model

import "time"

type WarReport struct {
	DB             dbSync    `xorm:"-"`
	Id             int       `xorm:"id pk autoincr"`
	AttackRid      int       `xorm:"attack_rid"`
	DefenseRid     int       `xorm:"defense_rid"`
	BegAttackArmy  string    `xorm:"beg_attack_army"`
	BegDefenseArmy string    `xorm:"beg_defense_army"`
	EndAttackArmy  string    `xorm:"end_attack_army"`
	EndDefenseArmy string    `xorm:"end_defense_army"`
	AttackIsWin    bool      `xorm:"attack_is_win"`
	AttackIsRead   bool      `xorm:"attack_is_read"`
	X              int       `xorm:"x"`
	Y              int       `xorm:"y"`
	Time           time.Time `xorm:"time"`
}

func (this *WarReport) TableName() string {
	return "war_report"
}

