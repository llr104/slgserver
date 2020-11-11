package http

type MyError struct {
	err string
	id  int
}

func New(err string, id int) error {
	return &MyError{err: err, id: id}
}

func (self *MyError) Error() string {
	return self.err
}

func (self *MyError) Id() int {
	return self.id
}



