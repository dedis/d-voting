package types

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	selectIDTest       = "select:"
	rankIDTest         = "rank:"
	textIDTest         = "text:"
	unmarshalingRankID = "could not unmarshal rank answers: "
	unmarshalingTextID = "could not unmarshal text answers: "
)

// Creating a ballot for the first question, which is a select question.
var ballot1 = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
	rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
	selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
	textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

// Creating a ballot with the following questions:
// 1. Select one of three options
// 2. Rank four options
// 3. Select one of five options
// 4. Write two text answers
// 5. Write one text answer
var ballot2 = string(selectIDTest + encodedQuestionID(1) + ":0,0,0\n" +
	rankIDTest + encodedQuestionID(2) + ":128,128,128,128\n" +
	selectIDTest + encodedQuestionID(3) + ":0,0,0,0,0\n" +
	textIDTest + encodedQuestionID(4) + ":xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx," +
	"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n" +
	textIDTest + encodedQuestionID(5) + ":xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,,\n\n")

func encodedQuestionID(i int) ID {
	return ID(base64.StdEncoding.EncodeToString([]byte("Q" + strconv.Itoa(i))))
}

func decodedQuestionID(i int) ID {
	return ID("Q" + strconv.Itoa(i))

}

func TestBallot_Unmarshal(t *testing.T) {
	b := Ballot{}
	form := Form{BallotSize: len(ballot1)}
	err := b.Unmarshal(ballot1, form)

	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	form.Configuration = Configuration{Scaffold: []Subject{{
		Subjects: []Subject{},

		Selects: []Select{{
			ID:      decodedQuestionID(1),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    2,
			MinN:    2,
			Choices: make([]string, 3),
		}, {
			ID:      decodedQuestionID(2),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    3,
			MinN:    3,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      decodedQuestionID(3),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        decodedQuestionID(4),
			Title:     {En: "", Fr: "", De: ""},
			MaxN:      2,
			MinN:      2,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}},
	},
	}}

	err = b.Unmarshal(ballot1, form)
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

	// with line wrongly formatted
	err = b.Unmarshal("x", form)
	require.EqualError(t, err, "a line in the ballot has length != 3: x")

	// with ID not encoded in base64
	ballotWrongID := string(selectIDTest + "aaa" + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongID, form)
	require.EqualError(t, err, "could not decode question ID: illegal base64 data at input byte 0")

	// with question ID not from the form
	ballotUnknownID := string(selectIDTest + encodedQuestionID(0) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotUnknownID, form)
	require.EqualError(t, err, "wrong question ID: the question doesn't exist")

	// with too many answers in select question
	ballotWrongSelect := string(selectIDTest + encodedQuestionID(1) + ":1,0,1,0,0\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, form)
	require.EqualError(t, err,
		"could not unmarshal select answers: question Q1 has a wrong number"+
			" of answers: expected 3 got 5")

	// with wrong format answers in select question
	ballotWrongSelect = string(selectIDTest + encodedQuestionID(1) + ":1,0,wrong\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, form)
	require.EqualError(t, err, "could not unmarshal select answers:"+
		" could not parse sform value for Q.Q1: strconv."+
		"ParseBool: parsing \"wrong\": invalid syntax")

	// with too many selected answers in select question
	ballotWrongSelect = string(selectIDTest + encodedQuestionID(1) + ":1,1,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, form)
	require.EqualError(t, err, "could not unmarshal select answers: "+
		"failed to check number of answers: question Q1 has too many selected answers")

	// with not enough selected answers in select question
	ballotWrongSelect = string(selectIDTest + encodedQuestionID(1) + ":1,0,0\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongSelect)

	err = b.Unmarshal(ballotWrongSelect, form)
	require.EqualError(t, err, "could not unmarshal select answers: "+
		"failed to check number of answers: question Q1 has not enough selected answers")

	// with not enough answers in rank question
	ballotWrongRank := string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	err = b.Unmarshal(ballotWrongRank, form)
	require.EqualError(t, err, "could not unmarshal rank answers: question"+
		" Q2 has a wrong number of answers: expected 5 got 3")

	// with wrong format answers in rank question
	ballotWrongRank = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,x,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, form)
	require.EqualError(t, err, unmarshalingRankID+
		"could not parse rank value for Q.Q2: strconv.ParseInt: parsing \"x\": invalid syntax")

	// with too many selected answers in rank question
	ballotWrongRank = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,3,4\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, form)
	require.EqualError(t, err, unmarshalingRankID+
		"invalid rank not in range [0, MaxN[: 3")

	// with valid ranks but one is selected twice
	ballotWrongRank = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,2,2\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, form)
	require.EqualError(t, err, unmarshalingRankID+
		"failed to check number of answers: question Q2 has too many selected answers")

	// with not enough selected answers in rank question
	ballotWrongRank = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongRank)

	err = b.Unmarshal(ballotWrongRank, form)
	require.EqualError(t, err, unmarshalingRankID+
		"failed to check number of answers: question"+
		" Q2 has not enough selected answers")

	// with not enough answers in text question
	ballotWrongText := string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, form)
	require.EqualError(t, err, unmarshalingTextID+
		"question Q4 has a wrong number of answers: expected 2 got 1")

	// with wrong encoding in text question
	ballotWrongText = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":wrongEncoding,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, form)
	require.EqualError(t, err, unmarshalingTextID+
		"could not decode text for Q.Q4: illegal base64 data at input byte 12")

	// with too many selected answers in text question
	form.Configuration.Scaffold[0].Texts[0].MaxN = 1

	ballotWrongText = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":YmxhYmxhYmxhZg==,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, form)
	require.EqualError(t, err, unmarshalingTextID+
		"failed to check number of answers: question Q4 has too many selected answers")

	form.Configuration.Scaffold[0].Texts[0].MaxN = 2

	// with not enough elected answers in text question
	ballotWrongText = string(selectIDTest + encodedQuestionID(1) + ":1,0,1\n" +
		rankIDTest + encodedQuestionID(2) + ":1,2,0,,\n" +
		selectIDTest + encodedQuestionID(3) + ":1,0,1,1\n" +
		textIDTest + encodedQuestionID(4) + ":,Y2VzdG1vaUVtaQ==\n\n")

	form.BallotSize = len(ballotWrongText)

	err = b.Unmarshal(ballotWrongText, form)
	require.EqualError(t, err, unmarshalingTextID+
		"failed to check number of answers: question Q4 has not enough selected answers")

	// with unknown question type
	ballotWrongType := string("wrong:" + encodedQuestionID(1) + ":")

	err = b.Unmarshal(ballotWrongType, form)
	require.EqualError(t, err, "question type is unknown")
}

func TestSubject_MaxEncodedSize(t *testing.T) {
	subject := Subject{
		Subjects: []Subject{{
			ID:       "",
			Title:    {En: "", Fr: "", De: ""},
			Order:    nil,
			Subjects: []Subject{},
			Selects:  []Select{},
			Ranks:    []Rank{},
			Texts:    []Text{},
		}},

		Selects: []Select{{
			ID:      encodedQuestionID(1),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    3,
			MinN:    0,
			Choices: make([]string, 3),
		}, {
			ID:      encodedQuestionID(2),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    5,
			MinN:    0,
			Choices: make([]string, 5),
		}},

		Ranks: []Rank{{
			ID:      encodedQuestionID(3),
			Title:   {En: "", Fr: "", De: ""},
			MaxN:    4,
			MinN:    0,
			Choices: make([]string, 4),
		}},

		Texts: []Text{{
			ID:        encodedQuestionID(4),
			Title:     {En: "", Fr: "", De: ""},
			MaxN:      2,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 2),
		}, {
			ID:        encodedQuestionID(5),
			Title:     {En: "", Fr: "", De: ""},
			MaxN:      1,
			MinN:      0,
			MaxLength: 10,
			Regex:     "",
			Choices:   make([]string, 3),
		}},
	}

	conf := Configuration{
		Title:    {En: "", Fr: "", De: ""},
		Scaffold: []Subject{subject},
	}

	size := conf.MaxBallotSize()

	require.Equal(t, len(ballot2), size)
	require.Equal(t, subject.MaxEncodedSize(), size)
}

func TestSubject_IsValid(t *testing.T) {
	mainSubject := &Subject{
		ID:       ID(base64.StdEncoding.EncodeToString([]byte("S1"))),
		Title:    {En: "", Fr: "", De: ""},
		Order:    []ID{},
		Subjects: []Subject{},
		Selects:  []Select{},
		Ranks:    []Rank{},
		Texts:    []Text{},
	}

	subSubject := &Subject{
		ID:       ID(base64.StdEncoding.EncodeToString([]byte("S2"))),
		Title:    {En: "", Fr: "", De: ""},
		Order:    []ID{},
		Subjects: []Subject{},
		Selects:  []Select{},
		Ranks:    []Rank{},
		Texts:    []Text{},
	}

	configuration := Configuration{
		Title:    {En: "", Fr: "", De: ""},
		Scaffold: []Subject{*mainSubject, *subSubject},
	}

	valid := configuration.IsValid()
	require.True(t, valid)

	// with double IDs

	mainSubject.ID = ID(base64.StdEncoding.EncodeToString([]byte("S1")))

	mainSubject.Selects = []Select{{
		ID:      encodedQuestionID(1),
		Title:   {En: "", Fr: "", De: ""},
		MaxN:    0,
		MinN:    0,
		Choices: make([]string, 0),
	}}

	mainSubject.Ranks = []Rank{{
		ID:      encodedQuestionID(1),
		Title:   {En: "", Fr: "", De: ""},
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
		Title:   {En: "", Fr: "", De: ""},
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
		Title:   {En: "", Fr: "", De: ""},
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
		Title:     {En: "", Fr: "", De: ""},
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
