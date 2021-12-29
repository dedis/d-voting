package types

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

var ballot1 = string("select:" + questionID(1) + ":1,0,1\n" +
	"rank:" + questionID(2) + ":1,2,0,,\n" +
	"select:" + questionID(3) + ":1,0,1,1\n" +
	"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

var ballot2 = string("select:" + questionID(1) + ":0,0,0\n" +
	"rank:" + questionID(2) + ":128,128,128,128\n" +
	"select:" + questionID(3) + ":0,0,0,0,0\n" +
	"text:" + questionID(4) + ":xxxxxxxxxxxxxxxx,xxxxxxxxxxxxxxxx\n" +
	"text:" + questionID(5) + ":xxxxxxxxxxxxxxxx,,\n\n")

func questionID(i int) ID {
	return ID(base64.StdEncoding.EncodeToString([]byte("Q" + strconv.Itoa(i))))
}

func TestBallot_Unmarshal(t *testing.T) {
	b := Ballot{}
	election := Election{BallotSize: len(ballot1)}
	err := b.Unmarshal(ballot1, election)

	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	election.Configuration = Configuration{Scaffold: []Subject{{
		Subjects: []Subject{},

		Selects: []Select{{
			ID:      questionID(1),
			Title:   "",
			MaxN:    2,
			MinN:    2,
			Choices: make([]string, 3),
		}, {
			ID:      questionID(2),
			Title:   "",
			MaxN:    3,
			MinN:    3,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      questionID(3),
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        questionID(4),
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
		SelectResultIDs: []ID{questionID(1), questionID(3)},
		SelectResult:    [][]bool{{true, false, true}, {true, false, true, true}},

		RankResultIDs: []ID{questionID(2)},
		RankResult:    [][]int8{{1, 2, 0, -1, -1}},

		TextResultIDs: []ID{questionID(4)},
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
	require.EqualError(t, err, "a line in the ballot has length != 3")

	// with ID not encoded in base64
	ballotWrongID := string("select:" + "aaa" + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongID, election)
	require.EqualError(t, err, "could not decode question ID: illegal base64 data at input byte 0")

	// with question ID not from the election
	ballotUnknownID := string("select:" + questionID(0) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotUnknownID, election)
	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	// with too many answers in select question
	ballotWrongSelect := string("select:" + questionID(1) + ":1,0,1,0,0\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "question UTE= has a wrong number of answers: expected 3 got 5")

	// with wrong format answers in select question
	ballotWrongSelect = string("select:" + questionID(1) + ":1,0,wrong\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "could not parse selection value for Q.UTE=: strconv."+
		"ParseBool: parsing \"wrong\": invalid syntax")

	// with too many selected answers in select question
	ballotWrongSelect = string("select:" + questionID(1) + ":1,1,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "question UTE= has too many selected answers")

	// with not enough selected answers in select question
	ballotWrongSelect = string("select:" + questionID(1) + ":1,0,0\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, election)
	require.EqualError(t, err, "question UTE= has not enough selected answers")

	// with not enough answers in rank question
	ballotWrongRank := string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "question UTI= has a wrong number of answers: expected 5 got 3")

	// with wrong format answers in rank question
	ballotWrongRank = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,x,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "could not parse rank value for Q.UTI= : strconv.ParseInt: parsing \"x\": invalid syntax")

	// with too many selected answers in rank question
	ballotWrongRank = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,3,4\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "question UTI= has too many selected answers")

	// with not enough selected answers in rank question
	ballotWrongRank = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, election)
	require.EqualError(t, err, "question UTI= has not enough selected answers")

	// with not enough answers in text question
	ballotWrongText := string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "question UTQ= has a wrong number of answers: expected 2 got 1")

	// with wrong encoding in text question
	ballotWrongText = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":wrongEncoding,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "could not decode text for Q. UTQ=: illegal base64 data at input byte 12")

	// with too many selected answers in text question
	election.Configuration.Scaffold[0].Texts[0].MaxN = 1

	ballotWrongText = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "question UTQ= has too many selected answers")

	election.Configuration.Scaffold[0].Texts[0].MaxN = 2

	// with not enough elected answers in text question
	ballotWrongText = string("select:" + questionID(1) + ":1,0,1\n" +
		"rank:" + questionID(2) + ":1,2,0,,\n" +
		"select:" + questionID(3) + ":1,0,1,1\n" +
		"text:" + questionID(4) + ":,Y2VzdG1vaUVtaQ==\n\n")

	election.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, election)
	require.EqualError(t, err, "question UTQ= has not enough selected answers")

	// with unknown question type
	ballotWrongType := string("wrong:" + questionID(1) + ":")

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
			ID:      "Q1",
			Title:   "",
			MaxN:    3,
			MinN:    0,
			Choices: make([]string, 3),
		}, {
			ID:      "Q2",
			Title:   "",
			MaxN:    5,
			MinN:    0,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      "Q3",
			Title:   "",
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        "Q4",
			Title:     "",
			MaxN:      2,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}, {
			ID:        "Q5",
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
		ID:      questionID(1),
		Title:   "",
		MaxN:    0,
		MinN:    0,
		Choices: make([]string, 0),
	}}

	mainSubject.Ranks = []Rank{{
		ID:      questionID(1),
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
		ID:      questionID(2),
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
		ID:      questionID(1),
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
		ID:        questionID(3),
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
	mainSubject.Order = []ID{questionID(1)}

	configuration.Scaffold = []Subject{*mainSubject}

	valid = configuration.IsValid()
	require.False(t, valid)

}
