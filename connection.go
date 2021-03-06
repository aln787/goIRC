package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type ConnectionStatus int

const (
	SocketConnected ConnectionStatus = iota
	UserPassSent
	UserNickSent
	UserUserInfoSent
	UserRegistered
)

type User struct {
	Nick     string
	Ident    string
	RealName string
	Conn     net.Conn
	Status   ConnectionStatus
	Host     string
}

//test added
func (u *User) GetHead() string {
	return fmt.Sprintf(":%s!%s@%v", u.Nick, u.Ident, u.Host)
	//return fmt.Sprintf(":%s!%s@127.0.0.1", u.Nick, u.Ident)
}

func handleConnection(conn net.Conn, buses map[string]*EventBus) {
	defer conn.Close()

	client := User{Status: UserPassSent, Conn: conn}
	//myIP := net.Conn.RemoteAddr().String()
	remoteA := fmt.Sprintf("%v", conn.RemoteAddr())
	localA := conn.LocalAddr()
	client.Host = remoteA
	fmt.Printf("Remote Address: %v\nLocal Address: %v\n", remoteA, localA)

	reader := bufio.NewReader(conn)

	commands := make(map[string]func(map[string]*EventBus, *User, string, string))
	commands["JOIN"] = handleJoin
	commands["TOPIC"] = handleTopic
	commands["PRIVMSG"] = handleMsg
	commands["NICK"] = handleNick
	commands["PART"] = handlePart
	commands["HELP"] = handleHelp
	commands["LIST"] = handleList
	commands["PING"] = handlePing
	commands["PONG"] = handlePong
	commands["USER"] = handleUser

	for {
		status, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		status = strings.TrimSpace(status)
		statLen := strings.Split(status, " ")

		// allows user to enter empty strings
		if len(status) == 0 {
			conn.Write([]byte(""))
			continue
		} else if len(statLen) < 2 {
			cmd := strings.SplitN(status, " ", 1)
			cmd[0] = strings.ToUpper(cmd[0])
			if _, ok := commands[cmd[0]]; ok {
				commands[cmd[0]](buses, &client, "", "")
			}
		} else {
			if client.Status < UserRegistered {
				regCmd := strings.SplitN(status, " ", 2)
				regCmd[0] = strings.ToUpper(regCmd[0])
				fmt.Println("-" + regCmd[0] + "-" + regCmd[1])
				switch regCmd[0] {
				case "PASS": //need to remove this at some point!
					client.Nick = regCmd[1]
					client.Ident = regCmd[1]
					client.RealName = regCmd[1]
					client.Host = remoteA

					buses[client.Nick] = &EventBus{subscribers: make(map[EventType][]Subscriber), channel: nil}
					buses[client.Nick].Subscribe(PrivMsg, &client)
					client.Status = UserRegistered
					sendWelcome(&client)

				//conn.Write([]byte("Welcome " + regCmd[1] + ")

				default:
					client.Write("you must register first. try nick or user?")
				}

			} else {
				// split <command> <target> :<data>

				var cmd, target, data string
				s := strings.SplitN(status, ":", 2)
				if len(s) > 1 {
					data = s[1]
				}
				_, err = fmt.Sscanf(s[0], "%s %s", &cmd, &target)
				if err != nil {
					fmt.Println(err)
					client.Write("Invalid input! CHECK YOUR(self) SYNTAX")
					continue
				}
				cmd = strings.ToUpper(cmd)
				if _, ok := commands[cmd]; ok {
					commands[cmd](buses, &client, target, data)
				}
			}
		}
	}
}

//test added, removed return param of the actual bus it appeared unneeded
func checkEventBus(buses map[string]*EventBus, client *User, target string) bool {
	_, ok := buses[target]
	if !ok {
		client.Write(fmt.Sprintf(canned_responses[ERR_NOSUCHCHANNEL], client.Nick))
	}
	return ok
}

//test added
func checkSubscribed(bus *EventBus, client *User, eventType EventType) bool {
	for _, v := range bus.subscribers[eventType] {
		if v == client {
			return true
		}
	}
	return false
}

//test added
func isChannel(target string) bool {
	return string(target[0]) == "#"
}

//test added
func handlePart(buses map[string]*EventBus, client *User, target string, data string) {
	message := fmt.Sprintf("%s parted %s!\n", client.Nick, target)

	if ok := checkEventBus(buses, client, target); !ok {
		return
	}
	if ok := checkSubscribed(buses[target], client, UserPart); !ok {
		return
	}
	buses[target].Publish(&Event{eventType: UserPart, event_data: message})
	delete(buses[target].channel.mode, client.Nick)

	buses[target].Unsubscribe(UserPart, client)
	buses[target].Unsubscribe(UserJoin, client)
	buses[target].Unsubscribe(Topic, client)
	buses[target].Unsubscribe(PrivMsg, client)
	// possibile race condition
	if len(buses[target].GetSubscribers(PrivMsg)) == 0 {
		delete(buses, target)
		fmt.Println(target + " closed")
	}
}

//test added - but it fails
func handlePing(buses map[string]*EventBus, client *User, target string, data string) {
	client.Write("PONG :" + target)
}

func handlePong(buses map[string]*EventBus, client *User, target string, data string) {
	//no op for fun
}

func handleJoin(buses map[string]*EventBus, client *User, target string, data string) {
	fmt.Println("!!!!!!!!! JOIN")
	if !isChannel(target) {
		return
	}
	var b *EventBus

	if ok := checkEventBus(buses, client, target); !ok {
		newChannel := Channel{name: target, topic: "gogo new channel!", mode: make(map[string]Mode)}
		buses[newChannel.name] = &EventBus{subscribers: make(map[EventType][]Subscriber), channel: &newChannel}
		b = buses[newChannel.name]
	} else {
		b = buses[target]
	}

	if ok := checkSubscribed(b, client, UserJoin); !ok {
		b.channel.mode[client.Nick] = Voice
		b.Subscribe(UserJoin, client)
		b.Subscribe(UserPart, client)
		b.Subscribe(PrivMsg, client)
		b.Subscribe(Topic, client)
		//message := fmt.Sprintf("%s joined %s!\n", client.Nick, target)
		message := fmt.Sprintf("%s JOIN %s\n", client.GetHead(), target)
		//send names
		var names string
		for _, val := range buses[target].subscribers[PrivMsg] {
			names = names + " " + val.GetInfo()
		}
		client.Write(":" + HOST_STRING + " 332 " + client.Nick + " " + target + ":no topic set")
		client.Write(":" + HOST_STRING + " 333 " + client.Nick + " " + target + " admin!admin@localhost 1419044230")
		client.Write(":" + HOST_STRING + " 353 " + client.Nick + " " + target + " :" + names)
		client.Write(":" + HOST_STRING + " 366 " + client.Nick + " * :END of /NAMES list.")
		///end send names
		b.Publish(&Event{UserJoin, message})
	}

}

func handleTopic(buses map[string]*EventBus, client *User, target string, data string) {

	if ok := checkEventBus(buses, client, target); !ok {
		return
	}
	if ok := checkSubscribed(buses[target], client, Topic); !ok {
		return
	}

	if len(data) > 0 {
		buses[target].channel.topic = data
		message := fmt.Sprintf("%s changed the channel topic to %s", client.Nick, data)
		buses[target].Publish(&Event{Topic, message})
	} else {
		message := fmt.Sprintf("%s\n", buses[target].channel.topic)
		client.Write(message)
	}
}

//test added
func handleNick(buses map[string]*EventBus, client *User, target string, data string) {
	client.Nick = target
	client.Write("nick set to:" + client.Nick)

	if client.Status == UserPassSent {
		client.Status = UserNickSent
	}
}

func handleMsg(buses map[string]*EventBus, client *User, target string, data string) {
	if ok := checkEventBus(buses, client, target); !ok {
		return
	}
	if !isChannel(target) {
		message := fmt.Sprintf("%s PRIVMSG %s: %s\n", client.GetHead(), target, data)
		buses[target].Publish(&Event{eventType: PrivMsg, event_data: message})
		buses[client.Nick].Publish(&Event{eventType: PrivMsg, event_data: message})
	}
	if ok := checkSubscribed(buses[target], client, PrivMsg); !ok {
		return
	}
	message := fmt.Sprintf("%s PRIVMSG %s: %s\n", client.GetHead(), target, data)
	buses[target].Publish(&Event{eventType: PrivMsg, event_data: message})
}

func handleList(buses map[string]*EventBus, client *User, target string, data string) {
	if len(buses) == 0 {
		client.Write("No Channels Exist")
	} else {
		client.Write("Channels")
		for k, _ := range buses {
			if k[:1] == "#" {
				client.Conn.Write([]byte(k + "\n"))
			}
		}
		client.Conn.Write([]byte("End of List\n"))
		//client.Write(k)
	}
}

//test added, but FIX PARSING SO WE ACTUALLY GET IDENT
func handleUser(buses map[string]*EventBus, client *User, target string, data string) {
	if client.Status != UserNickSent {
		//write an error to client, stating already registered
		return
	}

	fmt.Println("hit user case")
	var uname, hname, sname, rname string
	fmt.Sscanf(data, "%s %s %s :%s", uname, hname, sname, rname)
	fmt.Println(hname + uname)
	client.RealName = rname
	client.Status = UserRegistered
	client.Ident = client.Nick //terrible hack. fixme!
	fmt.Println("username:" + client.Ident)
	buses[client.Nick] = &EventBus{subscribers: make(map[EventType][]Subscriber), channel: nil}
	buses[client.Nick].Subscribe(PrivMsg, client)
	sendWelcome(client)
}

func handleHelp(buses map[string]*EventBus, client *User, target string, data string) {
	k, ok := Help[target]
	if !ok {
		client.Write("\nAvailable Commands: (Enter HELP <command> for further details")
		for h := range Help {
			client.Write(h)
		}
	} else {
		client.Write("Summary: " + k.Summary + "\r\nUsage: " + k.Syntax)
	}
}
