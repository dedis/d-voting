# Encoding of a ballot

Here is described a common encoding for ballots, which is needed
to make sure that all ballots can be encoded in a unique way (up to the
ordering of questions).

## Answers

The answers to questions are encoded in the following way, with one question per line:

```
<type><sep><id<sep><answers>

TYPE = "select"|"text"|"rank"
SEP = ":"
ID = 3 bytes, encoded in base64
ANSWERS = <answer>[","<answer>]*
ANSWER = <select_answer>|<text_answer>|<rank_answer>
SELECT_ANSWER = "0"|"1"
RANK_ANSWER = empty if not selected, or int in [0,MaxN]
TEXT_ANSWER = UTF-8 string encoded using base64
```

Here is an example:

For the following questions :

```bash
    Object 1: How did you find the provided material, from 1 (bad) to 5 (excellent) ?
    Choices: 1 () 2 () 3 () 4 () 5 ()

	Object 4.2: Rank the cafeteria
	Choices: () BC () SV () Parmentier

	Object 3: Who were the two best TAs ?
	Choices: TA1 [______] TA2 [______]
```

A possible encoding of an answer would be (by string concatenation):

```
"select:3fb2:0,0,0,1,0\n" +

"rank:19c7:0,1,2\n" +

"text:cd13:base64("Noémien"),base64("Pierluca")\n"
```

## Size of the ballot

In order to maintain complete voter anonymity and untraceability of ballots throughout the
voting process, it is important that all encrypted ballots have the same size. To this aim,
the form has an attribute called "BallotSize" which is the size
that all ballots should have before they're encrypted. Smaller ballots should therefore be
padded in order to reach this size. To denote the end of the ballot and the start of the padding,
we use an empty line (\n\n). For a ballot size of 117, our ballot from the previous example
would then become:

```
"select:3fb2:0,0,0,1,0\n" +

"rank:19c7:0,1,2\n" +

"text:cd13:base64("Noémien"),base64("Pierluca")\n\n" +

"ndtTx5uxmvnllH1T7NgLORuUWbN"
```

## Chunks of ballot

The encoded ballot must then be divided into chunks of 29 or less bytes since the maximum size supported by the kyber library for the encryption is of 29 bytes.

For the previous example we would then have 5 chunks, the first 4 would contain 29 bytes, while the last chunk would contain a single byte.
