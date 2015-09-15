package irc

import (
	"reflect"
	"testing"
)

var eventTests = []struct {
	// Event parsing
	Prefix, Cmd string
	Params        []string

	// Prefix parsing
	Name, User, Host string

	// Total output
	Expect string
	IsNil  bool

	// FromChannel
	FromChan bool
}{
	{
		IsNil: true,
	},
	{
		Expect: ":asd  :",
		IsNil:  true,
	},
	{
		Expect: ":A",
		IsNil:  true,
	},
	{
		Prefix: "server.kevlar.net",
		Cmd:    "PING",

		Name: "server.kevlar.net",

		Expect: ":server.kevlar.net PING\n",
	},
	{
		Prefix: "server.kevlar.net",
		Cmd:    "NOTICE",
		Params:   []string{"user", "*** This is a test"},

		Name: "server.kevlar.net",

		Expect: ":server.kevlar.net NOTICE user :*** This is a test\n",
	},
	{
		Prefix: "belakA!belakB@a.host.com",
		Cmd:    "PRIVMSG",
		Params:   []string{"#somewhere", "*** This is a test"},

		Name: "belakA",
		User: "belakB",
		Host: "a.host.com",

		Expect:   ":belakA!belakB@a.host.com PRIVMSG #somewhere :*** This is a test\n",
		FromChan: true,
	},
	{
		Prefix: "freenode",
		Cmd:    "005",
		Params:   []string{"starkbot", "CHANLIMIT=#:120", "MORE", "are supported by this server"},

		Name: "freenode",

		Expect: ":freenode 005 starkbot CHANLIMIT=#:120 MORE :are supported by this server\n",
	},
	{
		Prefix: "belakA!belakB@a.host.com",
		Cmd:    "PRIVMSG",
		Params:   []string{"&somewhere", "*** This is a test"},

		Name: "belakA",
		User: "belakB",
		Host: "a.host.com",

		Expect:   ":belakA!belakB@a.host.com PRIVMSG &somewhere :*** This is a test\n",
		FromChan: true,
	},
	{
		Prefix: "belakA!belakB@a.host.com",
		Cmd:    "PRIVMSG",
		Params:   []string{"belak", "*** This is a test"},

		Name: "belakA",
		User: "belakB",
		Host: "a.host.com",

		Expect: ":belakA!belakB@a.host.com PRIVMSG belak :*** This is a test\n",
	},
	{
		Prefix: "A",
		Cmd:    "B",
		Params:   []string{"C"},

		Name: "A",

		Expect: ":A B C\n",
	},
	{
		Prefix: "A@B",
		Cmd:    "C",
		Params:   []string{"D"},

		Name: "A",
		Host: "B",

		Expect: ":A@B C D\n",
	},
	{
		Cmd:    "B",
		Params:   []string{"C"},
		Expect: "B C\n",
	},
	{
		Prefix: "A",
		Cmd:    "B",
		Params:   []string{"C", "D"},

		Name: "A",

		Expect: ":A B C D\n",
	},
}

func TestParseEvent(t *testing.T) {
	for i, test := range eventTests {
		e := ParseEvent(test.Expect)
		if e == nil && !test.IsNil {
			t.Errorf("%d. Got nil for valid event", i)
		} else if e != nil && test.IsNil {
			t.Errorf("%d. Didn't get nil for invalid event", i)
		}

		if e == nil {
			continue
		}

		if test.Cmd != e.Command {
			t.Errorf("%d. command = %q, want %q", i, e.Command, test.Cmd)
		}
		if len(test.Params) != len(e.Params) {
			t.Errorf("%d. args = %v, want %v", i, e.Params, test.Params)
		} else {
			for j := 0; j < len(test.Params) && j < len(e.Params); j++ {
				if test.Params[j] != e.Params[j] {
					t.Errorf("%d. arg[%d] = %q, want %q", i, j, e.Params[j], test.Params[j])
				}
			}
		}
	}
}

func BenchmarkParseEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseEvent(eventTests[i%len(eventTests)].Prefix)
	}
}

func TestParsePrefix(t *testing.T) {
	for i, test := range eventTests {
		// TODO: Not sure if we should be skipping empty strings or handling them.
		if test.Prefix == "" {
			continue
		}

		pi := ParsePrefix(test.Prefix)
		if pi == nil {
			t.Errorf("%d. Got nil for valid identity", i)
			continue
		}
		if test.Name != pi.Name {
			t.Errorf("%d. name = %q, want %q", i, pi.Name, test.Name)
		}
		if test.User != pi.User {
			t.Errorf("%d. user = %q, want %q", i, pi.User, test.User)
		}
		if test.Host != pi.Host {
			t.Errorf("%d. host = %q, want %q", i, pi.Host, test.Host)
		}
	}
}

func BenchmarkParsePrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParsePrefix(eventTests[i%len(eventTests)].Expect)
	}
}

func TestEventTrailing(t *testing.T) {
	for i, test := range eventTests {
		if test.IsNil {
			continue
		}

		e := ParseEvent(test.Expect)
		tr := e.Trailing()
		if len(test.Params) < 1 {
			if tr != "" {
				t.Errorf("%d. trailing = %q, want %q", i, tr, "")
			}
		} else if tr != test.Params[len(test.Params)-1] {
			t.Errorf("%d. trailing = %q, want %q", i, tr, test.Params[len(test.Params)-1])
		}
	}
}

func TestEventFromChan(t *testing.T) {
	for i, test := range eventTests {
		if test.IsNil {
			continue
		}

		e := ParseEvent(test.Expect)
		if e.FromChannel() != test.FromChan {
			t.Errorf("%d. fromchannel = %v, want %v", i, e.FromChannel(), test.FromChan)
		}
	}
}

func TestEventCopy(t *testing.T) {
	for i, test := range eventTests {
		if test.IsNil {
			continue
		}

		e := ParseEvent(test.Expect)
		c := e.Copy()

		if !reflect.DeepEqual(e, c) {
			t.Errorf("%d. copy = %q, want %q", i, e, c)
		}

		if c.Prefix != nil {
			c.Prefix.Name += "junk"
			if reflect.DeepEqual(e, c) {
				t.Errorf("%d. copyidentity matched when it shouldn't", i)
			}
		}

		c.Params = append(c.Params, "junk")
		if reflect.DeepEqual(e, c) {
			t.Errorf("%d. copyargs matched when it shouldn't", i)
		}
	}
}

func TestEventString(t *testing.T) {
	for i, test := range eventTests {
		if test.IsNil {
			continue
		}

		e := ParseEvent(test.Expect)
		if e.String()+"\n" != test.Expect {
			t.Errorf("%d. %s did not match %s", i, e.String(), test.Expect)
		}
	}
}
