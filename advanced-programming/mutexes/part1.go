package main

import (
  //"fmt"
  "sync"
  "sync/atomic"
  //"time"
)

type Mtx struct {
  lock int32
  cond *sync.Cond
}


func main() {
  Naive()
  WithCond()
}

func WithCond() {
  t := 0
  var wg sync.WaitGroup
  m := newMtx()

  for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
      m.Lock()

      defer wg.Done()
      defer m.Unlock()
      t += 1
    }()
  }

  wg.Wait()
}

func Naive() {
  t := 0
  var wg sync.WaitGroup
  m := newMtx()

  for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
      m.NaiveLock()
      defer m.NaiveUnlock()
      defer wg.Done()
      t += 1
    }()
  }

  wg.Wait()

}

func newMtx() Mtx {
  m := sync.Mutex{}
  return Mtx{lock:0, cond: sync.NewCond(&m)}
}

func (m *Mtx) NaiveLock(){
  lockAcquired := atomic.CompareAndSwapInt32(&m.lock, 0, 1)
  for !lockAcquired {
    //time.Sleep(5)
    lockAcquired = atomic.CompareAndSwapInt32(&m.lock, 0, 1)
  }
}

func (m *Mtx) NaiveUnlock() {
  lockReleased := atomic.CompareAndSwapInt32(&m.lock, 1, 0)
  if !lockReleased {
    panic("Tried to unlock mtx which was already unlocked")
  }
}


func (m *Mtx) Lock(){
  m.cond.L.Lock()
  defer m.cond.L.Unlock()
  lockAcquired := atomic.CompareAndSwapInt32(&m.lock, 0, 1)
  for !lockAcquired {
    m.cond.Wait()
    lockAcquired = atomic.CompareAndSwapInt32(&m.lock, 0, 1)
  }
  //m.cond.Wait()
}

func (m *Mtx) Unlock() {
  m.cond.L.Lock()
  defer m.cond.L.Unlock()
  lockReleased := atomic.CompareAndSwapInt32(&m.lock, 1, 0)
  if !lockReleased {
    panic("Tried to unlock mtx which was already unlocked")
  }
  m.cond.Signal()
}
