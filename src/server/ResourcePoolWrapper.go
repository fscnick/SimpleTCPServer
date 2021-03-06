package main

type InitFunction func() (interface{}, error)

type ResourcePoolWrapper struct {
	size int
	conn chan interface{}
}

/**
Call the init function size times. If the init function fails during any call, then
the creation of the pool is considered a failure.
We call the same function size times to make sure each connection shares the same
state.
*/
func (p *ResourcePoolWrapper) InitPool(size int, initfn InitFunction) error {
	// Create a buffered channel allowing size senders
	p.conn = make(chan interface{}, size)
	for x := 0; x < size; x++ {
		conn, err := initfn()
		if err != nil {
			return err
		}

		// If the init function succeeded, add the connection to the channel
		p.conn <- conn
	}
	p.size = size
	return nil
}

func (p *ResourcePoolWrapper) GetResource() interface{} {

	select {
	case resource, ok := <- p.conn:
		if ok != true {
			// Channel closed
			return nil
		} 
		
		return resource

	default:
		// No resource ready.
		return nil
	}
	return nil
	//return <-p.conn
}

func (p *ResourcePoolWrapper) ReleaseResource(conn interface{}) {
	p.conn <- conn
}

func (p *ResourcePoolWrapper) getFreeRescourceCount() int {
	return len(p.conn)
}

func (p *ResourcePoolWrapper) getUsedRescourceCount() int {
	// TODO: This might be a good implementation. Make it better.
	return p.size - len(p.conn)
}
