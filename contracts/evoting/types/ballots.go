package types

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Ballot contains all information about a simple ballot
type Ballot struct {

	// SelectResult contains the result of each Select question. The result of a
	// select is a list of boolean that says for each choice if it has been
	// selected or not.  The ID slice is used to map a question ID to its index
	// in the SelectResult slice
	SelectResultIDs []ID
	SelectResult    [][]bool

	// RankResult contains the result of each Rank question. The result of a
	// rank question is the list of ranks for each choice. A choice that hasn't
	// been ranked will have a value < 0. The ID slice is used to map a question
	// ID to its index in the RankResult slice
	RankResultIDs []ID
	RankResult    [][]int8

	// TextResult contains the result of each Text question. The result of a
	// text question is the list of text answer for each choice. The ID slice is
	// used to map a question ID to its index in the TextResult slice
	TextResultIDs []ID
	TextResult    [][]string
}

// Unmarshal decodes the given string according to the format described in
// "state of smart contract.md"
func (b *Ballot) Unmarshal(marshalledBallot string, election Election) error {
	if len(marshalledBallot) > election.BallotSize {
		b.invalidate()
		return fmt.Errorf("ballot has an unexpected size %d, expected <= %d",
			len(marshalledBallot), election.BallotSize)
	}

	lines := strings.Split(marshalledBallot, "\n")

	b.SelectResultIDs = make([]ID, 0)
	b.SelectResult = make([][]bool, 0)

	b.RankResultIDs = make([]ID, 0)
	b.RankResult = make([][]int8, 0)

	b.TextResultIDs = make([]ID, 0)
	b.TextResult = make([][]string, 0)

	//TODO: Loads of code duplication, can be re-thought
	for _, line := range lines {
		question := strings.Split(line, ":")

		if len(question) != 3 {
			b.invalidate()
			return fmt.Errorf("a line in the ballot has length != 3")
		}

		q := election.Configuration.GetQuestion(ID(question[1]))

		if q == nil {
			b.invalidate()
			return fmt.Errorf("wrong question ID: the question doesn't exist")
		}

		switch question[0] {

		case "select":
			selections := strings.Split(question[2], ",")

			if len(selections) != q.GetChoicesLength() {
				b.invalidate()
				return fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", question[1], q.GetChoicesLength(), len(selections))
			}

			b.SelectResultIDs = append(b.SelectResultIDs, ID(question[1]))
			b.SelectResult = append(b.SelectResult, make([]bool, 0))

			index := len(b.SelectResult) - 1
			var selected uint = 0

			for _, selection := range selections {
				s, err := strconv.ParseBool(selection)

				if err != nil {
					b.invalidate()
					return fmt.Errorf("could not parse selection value for Q.%s: %v",
						question[1], err)
				}

				if s {
					selected += 1
				}

				b.SelectResult[index] = append(b.SelectResult[index], s)
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", question[1])
			}

		case "rank":
			ranks := strings.Split(question[2], ",")

			if len(ranks) != q.GetChoicesLength() {
				b.invalidate()
				return fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", question[1], q.GetChoicesLength(), len(ranks))
			}

			b.RankResultIDs = append(b.RankResultIDs, ID(question[1]))
			b.RankResult = append(b.RankResult, make([]int8, 0))

			index := len(b.RankResult) - 1
			var selected uint = 0
			for _, rank := range ranks {
				if len(rank) > 0 {
					selected += 1

					r, err := strconv.ParseInt(rank, 10, 8)
					if err != nil {
						b.invalidate()
						return fmt.Errorf("could not parse rank value for Q.%s : %v",
							question[1], err)
					}

					b.RankResult[index] = append(b.RankResult[index], int8(r))
				} else {
					b.RankResult[index] = append(b.RankResult[index], int8(-1))
				}
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", question[1])
			}

		case "text":
			texts := strings.Split(question[2], ",")

			if len(texts) != q.GetChoicesLength() {
				b.invalidate()
				return fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", question[1], q.GetChoicesLength(), len(texts))
			}

			b.TextResultIDs = append(b.TextResultIDs, ID(question[1]))
			b.TextResult = append(b.TextResult, make([]string, 0))

			index := len(b.TextResult) - 1
			var selected uint = 0

			for _, text := range texts {
				if len(text) > 0 {
					selected += 1
				}

				t, err := base64.StdEncoding.DecodeString(text)
				if err != nil {
					return fmt.Errorf("could not decode text for Q. %s", question[1])
				}

				b.TextResult[index] = append(b.TextResult[index], string(t))
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", question[1])
			}

		default:
			b.invalidate()
			return fmt.Errorf("question type is unknown")
		}

	}

	return nil
}

func (b *Ballot) invalidate() {
	b.RankResultIDs = nil
	b.RankResult = nil
	b.TextResultIDs = nil
	b.TextResult = nil
	b.SelectResultIDs = nil
	b.SelectResult = nil
}

// Subject is a wrapper around multiple questions that can be of type "select",
// "rank", or "text".
type Subject struct {
	ID ID

	Title string

	// Order defines the order of the different question, which all have a unique
	// identifier. This is purely for display purpose.
	Order []ID

	Subjects []Subject
	Selects  []Select
	Ranks    []Rank
	Texts    []Text
}

func (s *Subject) GetQuestion(ID ID) Question {
	for _, subject := range s.Subjects {
		question := subject.GetQuestion(ID)
		if question != nil {
			return question
		}
	}

	for _, selects := range s.Selects {
		if selects.ID == ID {
			return selects
		}
	}

	for _, rank := range s.Ranks {
		if rank.ID == ID {
			return rank
		}
	}

	for _, text := range s.Texts {
		if text.ID == ID {
			return text
		}
	}

	return nil
}

// MaxEncodedSize returns the maximum amount of bytes taken to store the
// questions in this subject once encoded in a ballot
func (s *Subject) MaxEncodedSize() int {
	size := 0

	for _, subject := range s.Subjects {
		size += subject.MaxEncodedSize()
	}

	for _, rank := range s.Ranks {
		size += len("rank::")
		size += len(rank.ID)
		// at most 4 bytes (-128) + ',' per choice
		size += len(rank.Choices) * 5
	}

	for _, selection := range s.Selects {
		size += len("select::")
		size += len(selection.ID)
		// 1 bytes (0/1) + ',' per choice
		size += len(selection.Choices) * 2
	}

	for _, text := range s.Texts {
		size += len("text::")
		size += len(text.ID)
		size += (int(text.MaxLength)+1)*int(text.MaxN) +
			int(math.Max(float64(len(text.Choices)-int(text.MaxN)), 0))
	}

	// Last line doesn't have a '\n'
	size -= 1

	return size
}

// IsValid verifies that all IDs are unique and the questions have coherent
// characteristics
func (s *Subject) IsValid(uniqueIDs map[ID]bool) bool {
	prevMapSize := len(uniqueIDs)

	uniqueIDs[s.ID] = true

	for _, rank := range s.Ranks {
		uniqueIDs[rank.ID] = true

		if !IsValid(rank) {
			return false
		}
	}

	for _, selection := range s.Selects {
		uniqueIDs[selection.ID] = true

		if !IsValid(selection) {
			return false
		}
	}

	for _, text := range s.Texts {
		uniqueIDs[text.ID] = true

		if !IsValid(text) {
			return false
		}
	}

	// If some ID was not unique
	currentMapSize := len(uniqueIDs)
	if prevMapSize+len(s.Ranks)+len(s.Texts)+len(s.Selects)+1 < currentMapSize {
		return false
	}

	for _, subject := range s.Subjects {
		if !subject.IsValid(uniqueIDs) {
			return false
		}
	}

	for _, id := range s.Order {
		exists := uniqueIDs[id]
		if !exists {
			return false
		}
	}

	return true
}

// Question is an interface offering the primitives all questions should have to
// verify the validity of an answer on a decrypted ballot.
type Question interface {
	GetMaxN() uint
	GetMinN() uint
	GetChoicesLength() int
}

func IsValid(q Question) bool {
	return (q.GetMinN() <= q.GetMaxN()) && (q.GetMaxN() <= uint(q.GetChoicesLength()))
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices. implements Question
type Select struct {
	ID ID

	Title   string
	MaxN    uint
	MinN    uint
	Choices []string
}

func (s Select) GetMaxN() uint {
	return s.MaxN
}

func (s Select) GetMinN() uint {
	return s.MinN
}

func (s Select) GetChoicesLength() int {
	return len(s.Choices)
}

// Rank describes a "rank" question, which requires the user to rank choices.
// implements Question
type Rank struct {
	ID ID

	Title   string
	MaxN    uint
	MinN    uint
	Choices []string
}

func (r Rank) GetMaxN() uint {
	return r.MaxN
}

func (r Rank) GetMinN() uint {
	return r.MinN
}

func (r Rank) GetChoicesLength() int {
	return len(r.Choices)
}

// Text describes a "text" question, which allows the user to enter free text.
// implements Question
type Text struct {
	ID ID

	Title     string
	MaxN      uint
	MinN      uint
	MaxLength uint
	Regex     string
	Choices   []string
}

func (t Text) GetMaxN() uint {
	return t.MaxN
}

func (t Text) GetMinN() uint {
	return t.MinN
}

func (t Text) GetChoicesLength() int {
	return len(t.Choices)
}
