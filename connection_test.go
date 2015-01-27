package main

import (
	"testing"
)

func TestHandleNick(t *testing.T) {
	futureNick := "afuturenickname"
	client := BuildUser()
	handleNick(nil, &client, futureNick, "")

	if client.Nick != futureNick {
		t.Errorf("client.Nick should be %q but is %q", futureNick, client.Nick)
	}
}

func TestIsChannel(t *testing.T) {
	aChannel := "#gophers"
	notAChannel := "gophers"

	if isChannel(aChannel) != true {
		t.Errorf("%q is not a valid channel according to isChannel(), it is.", aChannel)
	}

	if isChannel(notAChannel) == true {
		t.Errorf("%q is a valid channel according to isChannel(), it isn't.", notAChannel)
	}
}

func TestCheckEventBus(t *testing.T) {
	client := BuildUser()
	buses := make(map[string]*EventBus)
	key := "#gophers"
	randomKey := "#gjfkldfjglfdgfd"

	buses[key] = &EventBus{}

	if checkEventBus(buses, &client, key) != true {
		t.Errorf("checkEventBus states %q does not exist. It does.", key)
	}

	if checkEventBus(buses, &client, randomKey) == true {
		t.Errorf("checkEventBus states %q does exist, it doesn't.", randomKey)
	}
}

func TestCheckSubscribe(t *testing.T) {
	client := BuildUser()
	chanName := "#gophers"
	newChannel := Channel{name: chanName, topic: "gogo new channel!", mode: make(map[string]Mode)}
	buses := make(map[string]*EventBus)
	buses[newChannel.name] = &EventBus{subscribers: make(map[EventType][]Subscriber), channel: &newChannel}
	buses[newChannel.name].Subscribe(UserJoin, &client)
	checkSubscribed(buses[newChannel.name], &client, UserJoin)

	if checkSubscribed(buses[newChannel.name], &client, UserJoin) != true {
		t.Errorf("checkSubscribed says %q isn't subscribed to %q, it is.", client.Nick, newChannel.name)
	}

	if checkSubscribed(buses[newChannel.name], &client, UserPart) == true {
		t.Errorf("checkSubscribed says %q is subscribed, but it is not.", client.Nick)
	}
}
