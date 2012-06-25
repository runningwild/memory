Simple memory manager.  Useful in some cases where managing your own memory is simple, and where waiting for the GC to collect it and make it reusable causes unnecessary slowdown.

    manager := memory.NewManager()

    // Grab a block of a particular size
    block := manager.GetBlock(2000)

    // Give the block back to the manager so we can use it again later.
    manager.FreeBlock(block)
