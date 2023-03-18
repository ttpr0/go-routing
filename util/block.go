package util

type Block struct {
	channel   chan bool
	is_locked bool
}

func (self *Block) Take() {
	self.Release()
	self.is_locked = true
	<-self.channel
	self.is_locked = false
}
func (self *Block) Release() {
	if self.is_locked {
		self.channel <- true
	}
}
func (self *Block) IsTaken() bool {
	return self.is_locked
}

func NewBlock() *Block {
	return &Block{
		channel:   make(chan bool),
		is_locked: false,
	}
}
