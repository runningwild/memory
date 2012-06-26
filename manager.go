// In cases where Go's garbage collecter is unresponsive, or even broken, this
// allows you to easily manage blocks of memory.  This can be useful if you
// anticipate churning through lots of memory very quickly.  This package
// creates a default Manager that can be used through the package functions, 
// but you can create your own Manager if you want to eventually return the
// memory back to Go.
package memory

import (
  "fmt"
  "sync"
)

// We don't bother allocating blocks smaller than this.  This choice is
// somewhat arbitrary, but should probably be appropriate for most
// applications that want to manage their own memory.
const smallestBlockSize = 1024

// A Manager allocates memory as necessary and won't free it as long as it
// exists.  If you call GetBlock() then it should be followed by a FreeBlock()
//when you are done with it.
type Manager struct {
  mutex sync.Mutex

  // blocks[size] = a slice of blocks of length 2^(10+size)
  blocks [][][]byte

  // Map of used blocks.  The key is the address of the first byte in the
  // block.
  used map[*byte]bool
}

func NewManager() *Manager {
  var m Manager
  m.blocks = make([][][]byte, 21)
  m.used = make(map[*byte]bool)
  return &m
}

// Returns a slice of length n of the smallest block available that is at
// least that large.  Will reuse an existing block if possible, otherwise
// it will allocate a new one.
func (m *Manager) GetBlock(n int) []byte {
  m.mutex.Lock()
  defer m.mutex.Unlock()

  // Find the smallest block that can accomodate the request.
  c := smallestBlockSize
  s := 0
  for c < n {
    c *= 2
    s++
  }

  // Check if a block already exists of the appropriate size.
  for i := range m.blocks[s] {
    if !m.used[&m.blocks[s][i][0]] {
      m.used[&m.blocks[s][i][0]] = true
      for j := range m.blocks[s][i] {
        m.blocks[s][i][j] = 0
      }
      return m.blocks[s][i]
    }
  }

  // Otherwise allocate a new block to use.
  new_block := make([]byte, c)
  m.blocks[s] = append(m.blocks[s], new_block)
  m.used[&new_block[0]] = true
  return new_block[0:n]
}

// Returns a block back to the Manager so that it can be reused.  Panics if
// this block has already been freed, or if it was not allocated by this
// manager.
func (m *Manager) FreeBlock(b []byte) {
  m.mutex.Lock()
  defer m.mutex.Unlock()
  if _, ok := m.used[&b[0]]; !ok {
    panic("Tried to free an unused block")
  }
  delete(m.used, &b[0])
}

// Human-readable rundown on how many blocks have been reserved by this
// manager, and which ones are currently in use.
func (m *Manager) TotalAllocations() string {
  c := smallestBlockSize
  var ret string
  total_used := 0
  total_allocated := 0
  for s := range m.blocks {
    used := 0
    for i := range m.blocks[s] {
      if m.used[&m.blocks[s][i][0]] {
        used++
      }
    }
    if used > 0 {
      ret += fmt.Sprintf("%d bytes: %d/%d blocks in use.\n", c, used, len(m.blocks[s]))
    }
    total_used += used * c
    total_allocated += len(m.blocks[s]) * c
    c *= 2
  }
  ret += fmt.Sprintf("Total memory used/allocated: %d/%d\n", total_used, total_allocated)
  return ret
}

var manager *Manager

func init() {
  manager = NewManager()
}

// Returns a slice of length n of the smallest block available that is at
// least that large.  Will reuse an existing block if possible, otherwise
// it will allocate a new one.
func GetBlock(n int) []byte {
  return manager.GetBlock(n)
}

// Returns a block back to the Manager so that it can be reused.  Panics if
// this block has already been freed, or if it was not allocated by this
// manager.
func FreeBlock(b []byte) {
  manager.FreeBlock(b)
}

// Human-readable rundown on how many blocks have been reserved by this
// manager, and which ones are currently in use.
func TotalAllocations() string {
  return manager.TotalAllocations()
}
