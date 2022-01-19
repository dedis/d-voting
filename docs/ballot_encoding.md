# Encoding of a ballot


Here is described a common enconding for ballots, which is needed
to make sure that all ballots can be encoded in a unique way (up to the
ordering of questions).

## Answers
The answers to questions are encoded in the following way, with one question per line:

```
<type><sep><id<sep><answers>

TYPE = "select"|"text"|"rank"
SEP = ":"
ID = up to 3 bytes, encoded in base64
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

In order to maintain complete voter anonimity and untraceability of ballots throughout the 
election process, it is important that all encrypted ballots have the same size. To this aim, 
the election has an attribute called "BallotSize" (multiple of 29) which is the size 
that all ballots should have before they're encrypted. Smaller ballots should therefore be 
padded in order to reach this size. To denote the end of the ballot and the start of the padding,
we use an empty line (\n\n). For a ballot size of 116, our ballot from the previous example 
would then become:

```
"select:3fb2:false,false,false,true,false\n" +

"rank:19c7:0,1,2\n" + 

"text:cd13:base64("Noémien"),base64("Pierluca")\n\n" +

"olspoa1029ruxeqX129i0"
```
