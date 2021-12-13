package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var ballot1 = "select:aaa:1,0,1\n" +
	"rank:bbb:1,2,0,-128\n" +
	"select:ddd:true,false,true,true,true\n" +
	"text:ccc:YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n"

var ballot2 = "select:aaa:0,0,0\n" +
	"rank:bbb:-128,-128,-128,-128\n" +
	"select:ddd:0,0,0,0,0\n" +
	"text:ccc:blablablaf,cestmoiEmi\n" +
	"text:eee:aaaaaaaaaa,,\n\n"

func TestBallot_Unmarshal(t *testing.T) {
	b := Ballot{}
	election := Election{BallotSize: len(ballot1)}
	err := b.Unmarshal(ballot1, election)

	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	election.Configuration = Configuration{Scaffold: []Subject{{
		Subjects: []Subject{},

		Selects: []Select{{
			ID:      "aaa",
			Title:   "",
			MaxN:    3,
			MinN:    0,
			Choices: make([]string, 3),
		}, {
			ID:      "ddd",
			Title:   "",
			MaxN:    5,
			MinN:    0,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      "bbb",
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        "ccc",
			Title:     "",
			MaxN:      2,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}},
	},
	}}

	err = b.Unmarshal(ballot1, election)
	require.NoError(t, err)

	// expected ballot
	expected := Ballot{
		SelectResultIDs: []ID{"aaa", "ddd"},
		SelectResult:    [][]bool{{true, false, true}, {true, false, true, true, true}},

		RankResultIDs: []ID{"bbb"},
		RankResult:    [][]int8{{1, 2, 0, -128}},

		TextResultIDs: []ID{"ccc"},
		TextResult:    [][]string{{"blablablaf", "cestmoiEmi"}},
	}

	// check for equality
	require.Equal(t, b.SelectResultIDs, expected.SelectResultIDs)
	require.Equal(t, b.SelectResult, expected.SelectResult)
	require.Equal(t, b.RankResultIDs, expected.RankResultIDs)
	require.Equal(t, b.RankResult, expected.RankResult)
	require.Equal(t, b.TextResultIDs, expected.TextResultIDs)
	require.Equal(t, b.TextResult, expected.TextResult)

}

func TestSubject_MaxEncodedSize(t *testing.T) {
	subject := Subject{
		Subjects: []Subject{},

		Selects: []Select{{
			ID:      "aaa",
			Title:   "",
			MaxN:    3,
			MinN:    0,
			Choices: make([]string, 3),
		}, {
			ID:      "ddd",
			Title:   "",
			MaxN:    5,
			MinN:    0,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      "bbb",
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        "ccc",
			Title:     "",
			MaxN:      2,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}, {
			ID:        "eee",
			Title:     "",
			MaxN:      1,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 3),
		}},
	}

	conf := Configuration{
		MainTitle: "",
		Scaffold:  []Subject{subject},
	}

	size := conf.MaxBallotSize()

	require.Equal(t, len(ballot2), size)
	require.Equal(t, subject.MaxEncodedSize(), size)
}

//TODO Test configuration requirements
