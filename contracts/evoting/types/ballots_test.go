package types

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var ballot1 = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
	"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
	"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
	"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

var ballot2 = string("select:" + encodedQuestionID(1) + ":0,0,0\n" +
	"rank:" + encodedQuestionID(2) + ":128,128,128,128\n" +
	"select:" + encodedQuestionID(3) + ":0,0,0,0,0\n" +
	"text:" + encodedQuestionID(4) + ":xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx," +
	"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n" +
	"text:" + encodedQuestionID(5) + ":xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,,\n\n")

func encodedQuestionID(i int) ID {
	return ID(base64.StdEncoding.EncodeToString([]byte("Q" + strconv.Itoa(i))))
}

func decodedQuestionID(i int) ID {
	return ID("Q" + strconv.Itoa(i))

}

func TestBallot_Unmarshal(t *testing.T) {
	b := Ballot{}
	election := Election{BallotSize: len(ballot1)}
	err := b.Unmarshal(ballot1, election)

	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	election.Configuration = Configuration{Scaffold: []Subject{{
		Subjects: []Subject{},

		Selects: []Select{{
			ID:      decodedQuestionID(1),
			Title:   "",
			MaxN:    2,
			MinN:    2,
			Choices: make([]string, 3),
		}, {
			ID:      decodedQuestionID(2),
			Title:   "",
			MaxN:    3,
			MinN:    3,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      decodedQuestionID(3),
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        decodedQuestionID(4),
			Title:     "",
			MaxN:      2,
			MinN:      2,
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
		SelectResultIDs: []ID{decodedQuestionID(1), decodedQuestionID(3)},
		SelectResult:    [][]bool{{true, false, true}, {true, false, true, true}},

		RankResultIDs: []ID{decodedQuestionID(2)},
		RankResult:    [][]int8{{1, 2, 0, -1, -1}},

		TextResultIDs: []ID{decodedQuestionID(4)},
		TextResult:    [][]string{{"blablablaf", "cestmoiEmi"}},
	}

	// check for equality
	require.Equal(t, expected.SelectResultIDs, b.SelectResultIDs)
	require.Equal(t, expected.SelectResult, b.SelectResult)
	require.Equal(t, expected.RankResultIDs, b.RankResultIDs)
	require.Equal(t, expected.RankResult, b.RankResult)
	require.Equal(t, expected.TextResultIDs, b.TextResultIDs)
	require.Equal(t, expected.TextResult, b.TextResult)

	// with ballot too long
	err = b.Unmarshal(ballot1+"x", election)
	require.EqualError(t, err, "ballot has an unexpected size 102, expected <= 101")

	// with line wrongly formatted
	err = b.Unmarshal("x", election)
	require.EqualError(t, err, "a line in the ballot has length != 3: x")

	// with ID not encoded in base64
	ballotWrongID := string("select:" + "aaa" + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongID, election)
	require.EqualError(t, err, "could not decode question ID: illegal base64 data at input byte 0")

	// with question ID not from the election
	ballotUnknownID := string("select:" + encodedQuestionID(0) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotUnknownID, election)
	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	// with too many answers in select question
	ballotWrongSelect := string("select:" + encodedQuestionID(1) + ":1,0,1,0,0\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err,
		"could not unmarshal select answers: question Q1 has a wrong number"+
			" of answers: expected 3 got 5")

	// with wrong format answers in select question
	ballotWrongSelect = string("select:" + encodedQuestionID(1) + ":1,0,wrong\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "could not unmarshal select answers:"+
		" could not parse selection value for Q.Q1: strconv."+
		"ParseBool: parsing \"wrong\": invalid syntax")

	// with too many selected answers in select question
	ballotWrongSelect = string("select:" + encodedQuestionID(1) + ":1,1,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "could not unmarshal select answers: "+
		"question Q1 has too many selected answers")

	// with not enough selected answers in select question
	ballotWrongSelect = string("select:" + encodedQuestionID(1) + ":1,0,0\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "could not unmarshal select answers: "+
		"question Q1 has not enough selected answers")

	// with not enough answers in rank question
	ballotWrongRank := string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not unmarshal rank answers: question"+
		" Q2 has a wrong number of answers: expected 5 got 3")

	// with wrong format answers in rank question
	ballotWrongRank = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,x,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not unmarshal rank answers: "+
		"could not parse rank value for Q.Q2 : strconv.ParseInt: parsing \"x\": invalid syntax")

	// with too many selected answers in rank question
	ballotWrongRank = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,3,4\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not unmarshal rank answers: "+
		"invalid rank not in range [0, MaxN[")

	// with valid ranks but one is selected twice
	ballotWrongRank = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,2,2\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not unmarshal rank answers: "+
		"question Q2 has too many selected answers")

	// with not enough selected answers in rank question
	ballotWrongRank = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not unmarshal rank answers: question"+
		" Q2 has not enough selected answers")

	// with not enough answers in text question
	ballotWrongText := string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "could not unmarshal text answers: "+
		"question Q4 has a wrong number of answers: expected 2 got 1")

	// with wrong encoding in text question
	ballotWrongText = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":wrongEncoding,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "could not unmarshal text answers: "+
		"could not decode text for Q. Q4: illegal base64 data at input byte 12")

	// with too many selected answers in text question
	election.Configuration.Scaffold[0].Texts[0].MaxN = 1

	ballotWrongText = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "could not unmarshal text answers: "+
		"question Q4 has too many selected answers")

	election.Configuration.Scaffold[0].Texts[0].MaxN = 2

	// with not enough elected answers in text question
	ballotWrongText = string("select:" + encodedQuestionID(1) + ":1,0,1\n" +
		"rank:" + encodedQuestionID(2) + ":1,2,0,,\n" +
		"select:" + encodedQuestionID(3) + ":1,0,1,1\n" +
		"text:" + encodedQuestionID(4) + ":,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "could not unmarshal text answers: "+
		"question Q4 has not enough selected answers")

	// with unknown question type
	ballotWrongType := string("wrong:" + encodedQuestionID(1) + ":")

	err = b.Unmarshal(ballotWrongType, election)
	require.EqualError(t, err, "question type is unknown")
}

func TestSubject_MaxEncodedSize(t *testing.T) {
	subject := Subject{
		Subjects: []Subject{{
			ID:       "",
			Title:    "",
			Order:    nil,
			Subjects: []Subject{},
			Selects:  []Select{},
			Ranks:    []Rank{},
			Texts:    []Text{},
		}},

		Selects: []Select{{
			ID:      encodedQuestionID(1),
			Title:   "",
			MaxN:    3,
			MinN:    0,
			Choices: make([]string, 3),
		}, {
			ID:      encodedQuestionID(2),
			Title:   "",
			MaxN:    5,
			MinN:    0,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      encodedQuestionID(3),
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        encodedQuestionID(4),
			Title:     "",
			MaxN:      2,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}, {
			ID:        encodedQuestionID(5),
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

func TestSubject_IsValid(t *testing.T) {
	mainSubject := &Subject{
		ID:       ID(base64.StdEncoding.EncodeToString([]byte("S1"))),
		Title:    "",
		Order:    []ID{},
		Subjects: []Subject{},
		Selects:  []Select{},
		Ranks:    []Rank{},
		Texts:    []Text{},
	}

	subSubject := &Subject{
		ID:       ID(base64.StdEncoding.EncodeToString([]byte("S2"))),
		Title:    "",
		Order:    []ID{},
		Subjects: []Subject{},
		Selects:  []Select{},
		Ranks:    []Rank{},
		Texts:    []Text{},
	}

	configuration := Configuration{
		MainTitle: "",
		Scaffold:  []Subject{*mainSubject, *subSubject},
	}

	valid := configuration.IsValid()
	require.True(t, valid)

	// with wrongly ID not in base64
	mainSubject.ID = "zzz"

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with double IDs

	mainSubject.ID = ID(base64.StdEncoding.EncodeToString([]byte("S1")))

	mainSubject.Selects = []Select{{
		ID:      encodedQuestionID(1),
		Title:   "",
		MaxN:    0,
		MinN:    0,
		Choices: make([]string, 0),
	}}

	mainSubject.Ranks = []Rank{{
		ID:      encodedQuestionID(1),
		Title:   "",
		MaxN:    0,
		MinN:    0,
		Choices: make([]string, 0),
	}}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with invalid Rank question

	mainSubject.Ranks[0] = Rank{
		ID:      encodedQuestionID(2),
		Title:   "",
		MaxN:    0,
		MinN:    2,
		Choices: make([]string, 0),
	}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with invalid Select question

	mainSubject.Ranks = []Rank{}
	mainSubject.Selects[0] = Select{
		ID:      encodedQuestionID(1),
		Title:   "",
		MaxN:    1,
		MinN:    0,
		Choices: make([]string, 0),
	}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with invalid Text question

	mainSubject.Selects = []Select{}
	mainSubject.Texts = []Text{{
		ID:        encodedQuestionID(3),
		Title:     "",
		MaxN:      2,
		MinN:      4,
		MaxLength: 0,
		Regex:     "",
		Choices:   make([]string, 0),
	}}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with invalid sub subject

	subSubject.Texts = mainSubject.Texts
	mainSubject.Texts = []Text{}

	configuration.Scaffold = []Subject{*mainSubject}
	valid = configuration.IsValid()

	require.True(t, valid)

	mainSubject.Subjects = []Subject{*subSubject}
	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

	// with unknown ID in Order

	mainSubject.Subjects = []Subject{}
	mainSubject.Order = []ID{encodedQuestionID(1)}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)
}

func TestBallot_Equal(t *testing.T) {
	type check struct {
		ballot    Ballot
		other     Ballot
		assertion require.BoolAssertionFunc
	}

	table := []check{
		{
			Ballot{},
			Ballot{},
			require.True,
		},
		{
			Ballot{SelectResultIDs: []ID{"1"}},
			Ballot{},
			require.False,
		},
		{
			Ballot{SelectResultIDs: []ID{"1"}},
			Ballot{SelectResultIDs: []ID{"0"}},
			require.False,
		},
		{
			Ballot{SelectResultIDs: []ID{"1"}},
			Ballot{SelectResultIDs: []ID{"1"}},
			require.True,
		},
		{
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{true}},
			},
			Ballot{
				SelectResultIDs: []ID{"1"},
			},
			require.False,
		},
		{
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{true}},
			},
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{false}},
			},
			require.False,
		},
		{
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{true}},
			},
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{false, false}},
			},
			require.False,
		},
		{
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{true}},
			},
			Ballot{
				SelectResultIDs: []ID{"1"},
				SelectResult:    [][]bool{{true}},
			},
			require.True,
		},
		{
			Ballot{RankResultIDs: []ID{"1"}},
			Ballot{},
			require.False,
		},
		{
			Ballot{RankResultIDs: []ID{"1"}},
			Ballot{RankResultIDs: []ID{"0"}},
			require.False,
		},
		{
			Ballot{RankResultIDs: []ID{"1"}},
			Ballot{RankResultIDs: []ID{"1"}},
			require.True,
		},
		{
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{1}},
			},
			Ballot{
				RankResultIDs: []ID{"1"},
			},
			require.False,
		},
		{
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{1}},
			},
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{0}},
			},
			require.False,
		},
		{
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{1}},
			},
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{0, 0}},
			},
			require.False,
		},
		{
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{1}},
			},
			Ballot{
				RankResultIDs: []ID{"1"},
				RankResult:    [][]int8{{1}},
			},
			require.True,
		},
		{
			Ballot{TextResultIDs: []ID{"1"}},
			Ballot{},
			require.False,
		},
		{
			Ballot{TextResultIDs: []ID{"1"}},
			Ballot{TextResultIDs: []ID{"0"}},
			require.False,
		},
		{
			Ballot{TextResultIDs: []ID{"1"}},
			Ballot{TextResultIDs: []ID{"1"}},
			require.True,
		},
		{
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"0"}},
			},
			Ballot{TextResultIDs: []ID{"1"}},
			require.False,
		},
		{
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"1"}},
			},
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"0"}},
			},
			require.False,
		},
		{
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"1"}},
			},
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"0", "0"}},
			},
			require.False,
		},
		{
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"1"}},
			},
			Ballot{
				TextResultIDs: []ID{"1"},
				TextResult:    [][]string{{"1"}},
			},
			require.True,
		},
	}

	for _, e := range table {
		e.assertion(t, e.ballot.Equal(e.other))
	}
}
