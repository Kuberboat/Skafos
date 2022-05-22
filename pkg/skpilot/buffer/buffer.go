package buffer

import (
	"sync"

	"p9t.io/skafos/pkg/skpilot/client"
)

// SkBuffer is the base type for all the buffers in Skafos.
//
// When discoverer or SkPilot finds that something needs to be updated and send to SkAgent, it will
// write the updated data into SkBuffer. Messager reads SkBuffer at set intervals. When buffers
// are not empty, it will retrieve the data and send them to SkAgent.
//
// Each SkAgent has its own SkBuffer.
type SkBuffer interface {
	// LockBuffer locks an SkBuffer.
	LockBuffer()
	// UnlockBuffer unlocks an SkBuffer.
	UnlockBuffer()
	// IsEmpty returns whether the buffer for an SkAgent is empty.
	IsEmpty(agentAddr string) bool
	// ResetAgentBuffer clears the buffer for an SkAgent.
	ResetAgentBuffer(agentAddr string)
	// AcceptAgent sends the data in the buffer to an SkAgent. On success, it will clear the buffer.
	// Otherwise, the buffer will not be cleared, and the data will be sent next time this function
	// gets called.
	AcceptAgent(agentAddr string, cli *client.SkClient, wg *sync.WaitGroup)
	// BufferType returns the type of an SkBuffer.
	BufferType() string
}
