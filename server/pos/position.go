package pos

import "sync"

var RPMgr = RolePosMgr{
	posCaches: make(map[position]map[int]int),
}

type position struct {
	X		int
	Y		int
}

type RolePosMgr struct {
	mutex     sync.RWMutex
	posCaches map[position]map[int]int
	ridCaches map[int]position
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
	p := position{x, y}
	_, ok := this.posCaches[p]
	if ok == false {
		this.posCaches[p] = make(map[int]int)
	}
	this.posCaches[p][rid] = rid
}

func (this *RolePosMgr) GetCellRoleIds(x, y, width, height int) []int{
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	l := make([]int, 0)
	for i := x-width; i <= x+width; i++ {
		for j := y-height; j <= y+height; j++ {
			pos := position{x, y}
			r, ok := this.posCaches[pos]
			if ok {
				for _, v := range r {
					l = append(l, v)
				}
			}
		}
	}
	return l
}

