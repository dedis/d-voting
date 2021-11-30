package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var ballot1 = "select:aaa:true,false,true\n" +
	"rank:bbb:1,2,0,-128\n" +
	"select:ddd:true,false,true,true,true\n" +
	"text:ccc:blablablaf,cestmoiEmi"

var ballot2 = "select:aaa:false,false,false\n" +
	"rank:bbb:-128,-128,-128,-128\n" +
	"select:ddd:false,false,false,false,false\n" +
	"text:ccc:blablablaf,cestmoiEmi"

func TestBallot_Unmarshal(t *testing.T) {
	b := Ballot{}

	err := b.Unmarshal(ballot1, Election{})

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
