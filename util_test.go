package main

import (
	"net"
	"time"
)

type MockConn struct {
	LinesWritten    []string
	LastLineWritten string
}

func (c MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c MockConn) Write(b []byte) (n int, err error) {
	c.LastLineWritten = string(b)
	return 0, nil
}

func (c MockConn) Close() error {
	return nil
}

func (c MockConn) LocalAddr() net.Addr {
	return nil
}

func (c MockConn) RemoteAddr() net.Addr {
	return nil
}

func (c MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func BuildUser() User {
	client := User{}
	client.Nick = "arandomnickname"
	client.Ident = "arandomident"
	client.Conn = &MockConn{}

	return client
}

func BuildChannelAndBus() map[string]*EventBus {
	buses := make(map[string]*EventBus)

	chanName := "#gophers"
	newChannel := Channel{name: chanName, topic: "gogo new channel!", mode: make(map[string]Mode)}
	buses[newChannel.name] = &EventBus{subscribers: make(map[EventType][]Subscriber), channel: &newChannel}

	return buses
}

func BuildChannel() Channel {
	chanName := "#gophers"
	newChannel := Channel{name: chanName, topic: "gogo new channel!", mode: make(map[string]Mode)}
	return newChannel
}
