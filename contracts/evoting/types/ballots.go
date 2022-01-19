package types

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/xerrors"
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
		if line == "" {
			// empty line, the valid part of the ballot is over
			break
		}

		question := strings.Split(line, ":")

		if len(question) != 3 {
			b.invalidate()
			return xerrors.Errorf("a line in the ballot has length != 3")
		}

		_, err := base64.StdEncoding.DecodeString(question[1])
		if err != nil {
			return xerrors.Errorf("could not decode question ID: %v", err)
		}
		questionID := question[1]

		q := election.Configuration.GetQuestion(ID(questionID))

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
					"", questionID, q.GetChoicesLength(), len(selections))
			}

			b.SelectResultIDs = append(b.SelectResultIDs, ID(questionID))
			b.SelectResult = append(b.SelectResult, make([]bool, 0))

			index := len(b.SelectResult) - 1
			var selected uint = 0

			for _, selection := range selections {
				s, err := strconv.ParseBool(selection)

				if err != nil {
					b.invalidate()
					return fmt.Errorf("could not parse selection value for Q.%s: %v",
						questionID, err)
				}

				if s {
					selected += 1
				}

				b.SelectResult[index] = append(b.SelectResult[index], s)
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				b.invalidate()
				return fmt.Errorf("question %s has not enough selected answers", questionID)
			}

		case "rank":
			ranks := strings.Split(question[2], ",")

			if len(ranks) != q.GetChoicesLength() {
				b.invalidate()
				return fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", questionID, q.GetChoicesLength(), len(ranks))
			}

			b.RankResultIDs = append(b.RankResultIDs, ID(questionID))
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
							questionID, err)
					}

					if r < 0 || uint(r) >= q.GetMaxN() {
						b.invalidate()
						return fmt.Errorf("invalid rank not in range [0, MaxN[")
					}

					b.RankResult[index] = append(b.RankResult[index], int8(r))
				} else {
					b.RankResult[index] = append(b.RankResult[index], int8(-1))
				}
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				b.invalidate()
				return fmt.Errorf("question %s has not enough selected answers", questionID)
			}

		case "text":
			texts := strings.Split(question[2], ",")

			if len(texts) != q.GetChoicesLength() {
				b.invalidate()
				return fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", questionID, q.GetChoicesLength(), len(texts))
			}

			b.TextResultIDs = append(b.TextResultIDs, ID(questionID))
			b.TextResult = append(b.TextResult, make([]string, 0))

			index := len(b.TextResult) - 1
			var selected uint = 0

			for _, text := range texts {
				if len(text) > 0 {
					selected += 1
				}

				t, err := base64.StdEncoding.DecodeString(text)
				if err != nil {
					return fmt.Errorf("could not decode text for Q. %s: %v", questionID, err)
				}

				b.TextResult[index] = append(b.TextResult[index], string(t))
			}

			if selected > q.GetMaxN() {
				b.invalidate()
				return fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				b.invalidate()
				return fmt.Errorf("question %s has not enough selected answers", questionID)
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

// GetQuestion finds the question associated to a given ID and returns it
// Returns nil if no question found.
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

	//TODO : optimise by computing max size according to number of choices and maxN
	for _, rank := range s.Ranks {
		size += len("rank::")
		size += len(rank.ID)
		// at most 3 bytes (128) + ',' per choice
		size += len(rank.Choices) * 4
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

		maxTextPerAnswer := base64.StdEncoding.EncodedLen(int(text.MaxLength)) + 1
		size += maxTextPerAnswer*int(text.MaxN) +
			int(math.Max(float64(len(text.Choices)-int(text.MaxN)), 0))
	}

	// Last line has 2 '\n'
	if size != 0 {
		size += 1
	}

	return size
}

// isValid verifies that all IDs are unique and the questions have coherent
// characteristics
func (s *Subject) isValid(uniqueIDs map[ID]bool) bool {
	prevMapSize := len(uniqueIDs)

	uniqueIDs[s.ID] = true

	for _, rank := range s.Ranks {
		uniqueIDs[rank.ID] = true

		if !isValid(rank) {
			return false
		}
	}

	for _, selection := range s.Selects {
		uniqueIDs[selection.ID] = true

		if !isValid(selection) {
			return false
		}
	}

	for _, text := range s.Texts {
		uniqueIDs[text.ID] = true

		if !isValid(text) {
			return false
		}
	}

	// If some ID was not unique
	currentMapSize := len(uniqueIDs)
	if prevMapSize+len(s.Ranks)+len(s.Texts)+len(s.Selects)+1 > currentMapSize {
		return false
	}

	for _, subject := range s.Subjects {
		if !subject.isValid(uniqueIDs) {
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

func isValid(q Question) bool {
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

// GetMaxN implements Question
func (s Select) GetMaxN() uint {
	return s.MaxN
}

// GetMinN implements Question
func (s Select) GetMinN() uint {
	return s.MinN
}

// GetChoicesLength implements Question
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

// GetMaxN implements Question
func (r Rank) GetMaxN() uint {
	return r.MaxN
}

// GetMinN implements Question
func (r Rank) GetMinN() uint {
	return r.MinN
}

// GetChoicesLength implements Question
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

// GetMaxN implements Question
func (t Text) GetMaxN() uint {
	return t.MaxN
}

// GetMinN implements Question
func (t Text) GetMinN() uint {
	return t.MinN
}

// GetChoicesLength implements Question
func (t Text) GetChoicesLength() int {
	return len(t.Choices)
}
