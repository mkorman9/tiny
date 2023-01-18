package tinytcp

import (
	"net"
	"sync"
	"time"
)

type socketsList struct {
	head        *socketNode
	tail        *socketNode
	size        int
	maxSize     int
	m           sync.RWMutex
	socketsPool sync.Pool
	readersPool sync.Pool
	writersPool sync.Pool
	nodesPool   sync.Pool
}

type socketNode struct {
	socket *Socket
	prev   *socketNode
	next   *socketNode
}

func newSocketsList(maxSize int) *socketsList {
	return &socketsList{
		maxSize: maxSize,
		socketsPool: sync.Pool{
			New: func() any {
				return &Socket{}
			},
		},
		readersPool: sync.Pool{
			New: func() any {
				return &byteCountingReader{}
			},
		},
		writersPool: sync.Pool{
			New: func() any {
				return &byteCountingWriter{}
			},
		},
		nodesPool: sync.Pool{
			New: func() any {
				return &socketNode{}
			},
		},
	}
}

func (s *socketsList) New(connection net.Conn) *Socket {
	socket := s.newSocket(connection)

	if registered := s.registerSocket(socket); !registered {
		// instantly terminate the connection if it can't be added to the pool
		_ = socket.connection.Close()
		s.recycleSocket(socket)
		return nil
	}

	return socket
}

func (s *socketsList) Len() int {
	return s.size
}

func (s *socketsList) Cleanup() {
	s.m.Lock()
	defer s.m.Unlock()

	var node = s.head
	for node != nil {
		socket := node.socket
		next := node.next

		if socket.IsClosed() {
			switch node {
			case s.head:
				s.head = node.next
			case s.tail:
				s.tail = node.prev
				s.tail.next = nil
			default:
				node.prev.next = node.next
				node.next.prev = node.prev
			}

			s.recycleNode(node)
			s.recycleSocket(socket)
			s.size--
		}

		node = next
	}
}

func (s *socketsList) Copy() []*Socket {
	s.m.RLock()
	defer s.m.RUnlock()

	var list []*Socket
	for node := s.head; node != nil; node = node.next {
		if !node.socket.IsClosed() {
			list = append(list, node.socket)
		}
	}

	return list
}

func (s *socketsList) ExecRead(f func(head *socketNode)) {
	s.m.RLock()
	defer s.m.RUnlock()

	f(s.head)
}

func (s *socketsList) newSocket(connection net.Conn) *Socket {
	reader := s.readersPool.Get().(*byteCountingReader)
	reader.reader = connection

	writer := s.writersPool.Get().(*byteCountingWriter)
	writer.writer = connection

	socket := s.socketsPool.Get().(*Socket)
	socket.remoteAddress = parseRemoteAddress(connection)
	socket.connectedAt = time.Now()
	socket.connection = connection
	socket.reader = reader
	socket.writer = writer
	socket.byteCountingReader = reader
	socket.byteCountingWriter = writer

	return socket
}

func (s *socketsList) recycleSocket(socket *Socket) {
	socket.byteCountingReader.reader = nil
	socket.byteCountingReader.totalBytes = 0
	socket.byteCountingReader.currentBytes = 0
	s.readersPool.Put(socket.byteCountingReader)

	socket.byteCountingWriter.writer = nil
	socket.byteCountingWriter.totalBytes = 0
	socket.byteCountingWriter.currentBytes = 0
	s.writersPool.Put(socket.byteCountingWriter)

	socket.reset()
	s.socketsPool.Put(socket)
}

func (s *socketsList) registerSocket(socket *Socket) bool {
	s.m.Lock()
	defer s.m.Unlock()

	if s.maxSize >= 0 && s.size >= s.maxSize {
		return false
	}

	node := s.newNode(socket)
	if s.head == nil {
		s.head = node
		s.tail = node
	} else {
		s.tail.next = node
		node.prev = s.tail
		s.tail = node
	}

	s.size++

	return true
}

func (s *socketsList) newNode(socket *Socket) *socketNode {
	node := s.nodesPool.Get().(*socketNode)
	node.socket = socket
	return node
}

func (s *socketsList) recycleNode(node *socketNode) {
	node.socket = nil
	node.next = nil
	node.prev = nil
	s.nodesPool.Put(node)
}
