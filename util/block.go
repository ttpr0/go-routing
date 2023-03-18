package util

import "sync"

type Block struct {
	channel      chan bool
	is_locked    bool
	take_lock    sync.Mutex
	release_lock sync.Mutex
}

func (self *Block) Take() {
	self.take_lock.Lock()
	defer self.take_lock.Unlock()
	self.Release()
	self.is_locked = true
	<-self.channel
}
func (self *Block) Release() {
	self.release_lock.Lock()
	defer self.release_lock.Unlock()
	if self.is_locked {
		self.channel <- true
		self.is_locked = false
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
