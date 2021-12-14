# Encoding of a ballot


Here is described a common enconding for ballots, which is needed
to make sure that all ballots can be encoded in a unique way (up to the
ordering of questions).

## Answers
The answers to questions are encoded in the following way:

```
type:ID:<answers>
```
Where we have that:
* the type can either be "select", "rank" or "text"  
* ID is the ID of the question 
* The 'answers' are seperated by commas
* There should be as many answers as there are "choices" in the question (even if empty). 
* There should be one question per line. 
* Answers to "select" questions should be '0' (false) or '1' (true)
* Answers to "text" questions should be encoded in base64.
* Answers to "rank" questions should be empty if not selected and the rank should only vary from 
  0 to the maximum amount of choices selected.

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
"select:0x3fb2:0,0,0,1,0\n" +

"rank:0x19c7:0,1,2\n" + 

"text:0xcd13:base64("Noémien"),base64("Pierluca")"
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
"select:0x3fb2:false,false,false,true,false\n" +

"rank:0x19c7:0,1,2\n" + 

"text:0xcd13:"Noémien","Pierluca"\n\n" +

"olspoa1029ruxeqX129i0"
```