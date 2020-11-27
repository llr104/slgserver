package pos

import "sync"

var RPMgr = RolePosMgr{
	posCaches: make(map[Position]map[int]int),
}

type Position struct {
	X		int
	Y		int
}

type RolePosMgr struct {
	mutex     sync.RWMutex
	posCaches map[Position]map[int]int
	ridCaches map[int]Position
}

func (this *RolePosMgr) Push(x, y, rid int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	//旧的要删除
	if r, ok := this.ridCaches[rid]; ok {
		if r.X == x && r.Y == y{
			return
		}
		if c, ok1 := this.posCaches[r]; ok1 {
			delete(c, rid)
		}
	}

	//新的写入
	p := Position{x, y}
	_, ok := this.posCaches[p]
	if ok == false {
		this.posCaches[p] = make(map[int]int)
	}
	this.posCaches[p][rid] = rid
}



