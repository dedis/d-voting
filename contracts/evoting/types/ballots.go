package types

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

const (
	selectID = "select"
	rankID   = "rank"
	textID   = "text"
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
func (b *Ballot) Unmarshal(marshalledBallot string, form Form) error {
	lines := strings.Split(marshalledBallot, "\n")

	b.SelectResultIDs = make([]ID, 0)
	b.SelectResult = make([][]bool, 0)

	b.RankResultIDs = make([]ID, 0)
	b.RankResult = make([][]int8, 0)

	b.TextResultIDs = make([]ID, 0)
	b.TextResult = make([][]string, 0)

	for _, line := range lines {
		if line == "" {
			// empty line, the valid part of the ballot is over
			break
		}

		question := strings.Split(line, ":")

		if len(question) != 3 {
			b.invalidate()
			return xerrors.Errorf("a line in the ballot has length != 3: %s", line)
		}

		questionID, err := base64.StdEncoding.DecodeString(question[1])
		if err != nil {
			return xerrors.Errorf("could not decode question ID: %v", err)
		}

		q := form.Configuration.GetQuestion(ID(questionID))

		if q == nil {
			b.invalidate()
			return fmt.Errorf("wrong question ID: the question doesn't exist")
		}

		switch question[0] {

		case selectID:
			selections := strings.Split(question[2], ",")

			selectQ := Select{
				ID:      ID(questionID),
				MaxN:    q.GetMaxN(),
				MinN:    q.GetMinN(),
				Choices: make([]string, q.GetChoicesLength()),
			}

			results, err := selectQ.unmarshalAnswers(selections)
			if err != nil {
				b.invalidate()
				return fmt.Errorf("could not unmarshal select answers: %v", err)
			}

			b.SelectResultIDs = append(b.SelectResultIDs, ID(questionID))
			b.SelectResult = append(b.SelectResult, results)

		case rankID:
			ranks := strings.Split(question[2], ",")

			rankQ := Rank{
				ID:      ID(questionID),
				MaxN:    q.GetMaxN(),
				MinN:    q.GetMinN(),
				Choices: make([]string, q.GetChoicesLength()),
			}

			results, err := rankQ.unmarshalAnswers(ranks)
			if err != nil {
				b.invalidate()
				return fmt.Errorf("could not unmarshal rank answers: %v", err)
			}
			b.RankResultIDs = append(b.RankResultIDs, ID(questionID))
			b.RankResult = append(b.RankResult, results)

		case textID:
			texts := strings.Split(question[2], ",")

			textQ := Text{
				ID:        ID(questionID),
				MaxN:      q.GetMaxN(),
				MinN:      q.GetMinN(),
				MaxLength: 0, // TODO: Should the length check be also done at decryption?
				Choices:   make([]string, q.GetChoicesLength()),
			}

			results, err := textQ.unmarshalAnswers(texts)
			if err != nil {
				b.invalidate()
				return fmt.Errorf("could not unmarshal text answers: %v", err)
			}
			b.TextResultIDs = append(b.TextResultIDs, ID(questionID))
			b.TextResult = append(b.TextResult, results)

		default:
			b.invalidate()
			return fmt.Errorf("question type is unknown")
		}

	}

	return nil
}

// checkNumberOfAnswers checks if the given amount of answers is in the accepted
// range for the given question
func checkNumberOfAnswers(maxN uint, minN uint, nbrOfAnswers uint, questionID ID) error {
	if nbrOfAnswers > maxN {
		return fmt.Errorf("question %s has too many selected answers", questionID)
	}
	if nbrOfAnswers < minN {
		return fmt.Errorf("question %s has not enough selected answers", questionID)
	}
	return nil
}

// invalidate makes the ballot invalid by putting all field to nil
func (b *Ballot) invalidate() {
	b.RankResultIDs = nil
	b.RankResult = nil
	b.TextResultIDs = nil
	b.TextResult = nil
	b.SelectResultIDs = nil
	b.SelectResult = nil
}

// Equal performs a loose comparison of a ballot.
func (b *Ballot) Equal(other Ballot) bool {
	if len(b.SelectResultIDs) != len(other.SelectResultIDs) {
		return false
	}

	for i, id := range b.SelectResultIDs {
		if id != other.SelectResultIDs[i] {
			return false
		}
	}

	if len(b.SelectResult) != len(other.SelectResult) {
		return false
	}

	for i, sr := range b.SelectResult {
		if len(sr) != len(other.SelectResult[i]) {
			return false
		}

		for j, r := range sr {
			if r != other.SelectResult[i][j] {
				return false
			}
		}
	}

	if len(b.RankResultIDs) != len(other.RankResultIDs) {
		return false
	}

	for i, id := range b.RankResultIDs {
		if id != other.RankResultIDs[i] {
			return false
		}
	}

	if len(b.RankResult) != len(other.RankResult) {
		return false
	}

	for i, rr := range b.RankResult {
		if len(rr) != len(other.RankResult[i]) {
			return false
		}

		for j, r := range rr {
			if r != other.RankResult[i][j] {
				return false
			}
		}
	}

	if len(b.TextResultIDs) != len(other.TextResultIDs) {
		return false
	}

	for i, id := range b.TextResultIDs {
		if id != other.TextResultIDs[i] {
			return false
		}
	}

	if len(b.TextResult) != len(other.TextResult) {
		return false
	}

	for i, tr := range b.TextResult {
		if len(tr) != len(other.TextResult[i]) {
			return false
		}

		for j, r := range tr {
			if r != other.TextResult[i][j] {
				return false
			}
		}
	}

	return true
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
		size += len(rank.GetID() + "::")
		size += len(rank.ID)
		// at most 3 bytes (128) + ',' per choice
		size += len(rank.Choices) * 4
	}

	for _, selection := range s.Selects {
		size += len(selection.GetID() + "::")
		size += len(selection.ID)
		// 1 bytes (0/1) + ',' per choice
		size += len(selection.Choices) * 2
	}

	for _, text := range s.Texts {
		size += len(text.GetID() + "::")
		size += len(text.ID)

		// at most 4 bytes per character + ',' per answer
		maxTextPerAnswer := 4*int(text.MaxLength) + 1
		size += maxTextPerAnswer*int(text.MaxN) +
			int(math.Max(float64(len(text.Choices)-int(text.MaxN)), 0))
	}

	// Last line has 2 '\n'
	if size != 0 {
		size++
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

	for _, sform := range s.Selects {
		uniqueIDs[sform.ID] = true

		if !isValid(sform) {
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

// Question is an offering the primitives all questions should have to
// verify the validity of an answer on a decrypted ballot.
type Question interface {
	GetMaxN() uint
	GetMinN() uint
	GetChoicesLength() int
	GetID() string
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
	Hint    string
}

// GetID implements Question
func (s Select) GetID() string {
	return selectID
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

// unmarshalAnswers interprets the given raw answers into a slice of bool with
// the answer for each choice and ensure the answers are correctly formatted
func (s Select) unmarshalAnswers(sforms []string) ([]bool, error) {
	if len(sforms) != len(s.Choices) {
		return nil, fmt.Errorf("question %s has a wrong number of answers:"+
			" expected %d got %d", s.ID, len(s.Choices), len(sforms))
	}

	var selected uint = 0
	results := make([]bool, 0)

	for _, sform := range sforms {
		b, err := strconv.ParseBool(sform)

		if err != nil {
			return nil, fmt.Errorf("could not parse sform value for Q.%s: %v",
				s.ID, err)
		}

		if b {
			selected++
		}

		results = append(results, b)
	}

	err := checkNumberOfAnswers(s.MaxN, s.MinN, selected, s.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check number of answers: %v", err)
	}

	return results, nil
}

// Rank describes a "rank" question, which requires the user to rank choices.
// implements Question
type Rank struct {
	ID ID

	Title   string
	MaxN    uint
	MinN    uint
	Choices []string
	Hint    string
}

func (r Rank) GetID() string {
	return rankID
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

// unmarshalAnswers interprets the given raw answers into a slice of integer
// representing the ranking of each choice and ensures the answers are correctly
// formatted
func (r Rank) unmarshalAnswers(ranks []string) ([]int8, error) {
	if len(ranks) != len(r.Choices) {
		return nil, fmt.Errorf("question %s has a wrong number of answers:"+
			" expected %d got %d", r.ID, len(r.Choices), len(ranks))
	}

	var selected uint = 0
	results := make([]int8, 0, len(ranks))

	for _, rank := range ranks {
		if len(rank) <= 0 {
			results = append(results, int8(-1))
			continue
		}

		selected++

		rankValue, err := strconv.ParseInt(rank, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("could not parse rank value for Q.%s: %v",
				r.ID, err)
		}

		if rankValue < 0 || uint(rankValue) >= r.MaxN {
			return nil, fmt.Errorf("invalid rank not in range [0, MaxN[: %d",
				rankValue)
		}

		results = append(results, int8(rankValue))
	}

	err := checkNumberOfAnswers(r.MaxN, r.MinN, selected, r.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check number of answers: %v", err)
	}

	return results, nil
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
	Hint      string
}

func (t Text) GetID() string {
	return textID
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

// unmarshalAnswers interprets the given raw answers into a slice with the
// decoded answer corresponding to each choice and ensure the answers are
// correctly formatted
func (t Text) unmarshalAnswers(texts []string) ([]string, error) {
	if len(texts) != len(t.Choices) {
		return nil, fmt.Errorf("question %s has a wrong number of answers:"+
			" expected %d got %d", t.ID, len(t.Choices), len(texts))
	}

	var selected uint = 0
	results := make([]string, 0)

	for _, text := range texts {
		if len(text) > 0 {
			selected++
		}

		textValue, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, fmt.Errorf("could not decode text for Q.%s: %v", t.ID, err)
		}

		results = append(results, string(textValue))
	}

	err := checkNumberOfAnswers(t.MaxN, t.MinN, selected, t.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check number of answers: %v", err)
	}

	return results, nil
}
