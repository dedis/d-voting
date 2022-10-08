# State of Smart Contract

In the use cases we defined two smart contracts for each of the following purposes:

- storing the forms informations
- storing a ballot

As (at least for the moment) in a dela there is no notion of “instance of a smart contract”, we have
to see the definition of a smart contract as a single entity that performs its read/write access on
the storage with a predefined set of keys.

The simplest and naive solution is to have a single smart contract that stores everything: the
forms, the ballots for each form, and the results. This is like having a giant XML/JSON file
that contains everything, with the smart contract being the interface that handles that.

As we want to be generic, we should be able to create “any” kind of poll. With that regard we
identified 3 types of question that can be asked in a poll:

- “Select” question
- “Rank” question
- “Text” question

**Select**: a select question asks the user to select among multiple choices, from 0 to N, where N
is the total number of choices. That kind of question can be “Select candidate A or B”, or “Select
two candidates from A, B, and C”.

**Rank**: a rank question asks the user to rank 0 or N choices, where N is the total number of
choices. That kind of question can be “Rank each car from the list based on your preference,
starting with 0, your favorite, to N”.

**Text**: a text question asks the user to enter a free text in 0 to N field, where N is the total
number of choices. That kind of question can be “Enter the name of 1 or 2 of your favorite
candidate(s)”.

The data structure of an form should contain the **Configuration**, which describes the
different questions that the poll contains. We gather the questions (ie. the collections of Select,
Rank, Text questions) under “Subjects” that define the order of the questions. A subject is nothing
more than a container to organize the questions. Each subject can also contain, apart from the
questions, other subjects. That allows us to have a flexible structure to describe nested
subsections of questions.

The answer of a poll given by a client is defined by its list of answers for each kind of question:
A list of **SelectAnswers**, **RankAnswers**, and
**TextAnswers**. Each question defined in the configuration of a poll has a unique identifier, which
allows us to map answers to the question.

```go
type ID uint16
type Identity uint16

type Forms struct {
Forms map[ID]Form
}

type Form struct {
    Configuration       Configuration
    Status              status // Initial | Open | Closed | Shuffling | Decrypting | ..
    Pubkey              []byte
    PublicBulletinBoard PublicBulletinBoard
    ShuffleInstances    []ShuffleInstance
    DecryptedBallots    []Ballot
}

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

// Configuration contains the configuration of a new poll.
type Configuration struct {
    MainTitle string
    Scaffold  []Subject
}

// Subject is a wrapper around multiple questions that can be of type "select",
// "rank", or "text".
type Subject struct {
    ID ID

    Title string

    // Order defines the order of the different question, which all have a uniq
    // identifier. This is purely for display purpose.
    Order []ID

    Subjects []Subject
    Selects  []Select
    Ranks    []Rank
    Texts    []Text
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices.
type Select struct {
    ID ID

    Title   string
    MaxN    int
    MinN    int
    Choices []string
}

// Rank describes a "rank" question, which requires the user to rank choices.
type Rank struct {
    ID ID

    Title   string
    MaxN    int
    MinN    int
    Choices []string
}

// Text describes a "text" question, which allows the user to enter free text.
type Text struct {
    ID int

    Title      string
    MaxN       int
    MinN       int
    MaxLength  int
    Regex      string
    Choices    []string
}
```

Here is an example of a poll we could want to run:

```bash
Voting title: Please give your opinion

Subject1: Rate the course

	Object 1: How did you find the provided material, from 1 (bad) to 5 (excellent) ?
	Choices: 1 () 2 () 3 () 4 () 5 ()

	Object 2: How did you find the teaching ?
	Choices: bad () normal () good ()

	Object 3: Who were the two best TAs ?
	Choices: TA1 [______] TA2 [______]

 	Subject 1.2:

		Object 4.1: Select your ingredients
		Choices: [] tomato [] salad [] onion

		Object 4.2: Rank the cafeteria
		Choices: () BC () SV () Parmentier

			...

Subject2:
	...
```

And here is the corresponding Configuration:

```bash
v := Configuration{
	MainTitle: "Please give your opinion",
	Scaffold: []Subject{
		{
			Title: "Rate the course",
			ID:    0xa2ab,
			Order: []ID{0x3fb2, 0x41e2, 0xcd13, 0xff31},

			Subjects: []Subject{
				{
					Title: "Let's talk about the food",
					ID:    0xff31,
					Order: []ID{0xa319, 0x19c7},

					Selects: []Select{
						Select{
							Title:   "Select your ingredients",
							ID:      0xa319,
							MaxN:    3,
							MinN:    0,
							Choices: []string{"tomato", "salad", "onion"},
						},
					},

					Ranks: []Rank{
						Rank{
							Title:   "Rank the cafeteria",
							ID:      0x19c7,
							MaxN:    3,
							MinN:    3,
							Choices: []string{"BC", "SV", "Parmentier"},
						},
					},
				},
			},

			Selects: []Select{
				Select{
					Title:   "How did you find the provided material, from 1 (bad) to 5 (excellent) ?",
					ID:      0x3fb2,
					MaxN:    1,
					MinN:    1,
					Choices: []string{"1", "2", "3", "4", "5"},
				},
				Select{
					Title:   "How did you find the teaching ?",
					ID:      0x41e2,
					MaxN:    1,
					MinN:    1,
					Choices: []string{"bad", "normal", "good"},
				},
			},

			Texts: []Text{
				Text{
					Title:     "Who were the two best TAs ?",
					ID:        0xcd13,
					MaxLength: 20,
					MaxN:      2,
					MinN:      1,
					Choices:   []string{"TA1", "TA2"},
				},
			},
		},
	},
}
```
