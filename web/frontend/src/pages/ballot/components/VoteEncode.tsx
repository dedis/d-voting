import { Buffer } from 'buffer';
import ShortUniqueId from 'short-unique-id';
import { Answers, RANK, SELECT, TEXT } from 'types/configuration';

export function voteEncode(answers: Answers, chunksPerBallot: number): string[] {
  // contains the special string representation of the result
  let encodedBallot = '';

  answers.SelectAnswers.forEach((selectAnswer, id) => {
    encodedBallot += SELECT + ':' + id + ':';
    selectAnswer.forEach((answer) => (encodedBallot += answer ? '1,' : '0,'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.RankAnswers.forEach((rankAnswer, id) => {
    encodedBallot += RANK + ':' + id + ':';
    const position = Array<number>(rankAnswer.length);
    for (let i = 0; i < rankAnswer.length; i++) {
      position[rankAnswer[i]] = i;
    }
    position.forEach((pos) => (encodedBallot += pos + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.TextAnswers.forEach((textAnswer, id) => {
    encodedBallot += TEXT + ':' + id + ':';
    // each answer is first transformed into bytes then encoded in base64
    textAnswer.forEach((answer) => (encodedBallot += Buffer.from(answer).toString('base64') + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  encodedBallot += '\n';

  const chunkSize = 29;
  // ballot size
  const ballotSize = chunksPerBallot * chunkSize;
  const encodedBallotSize = Buffer.byteLength(encodedBallot);

  // add padding if necessary
  if (encodedBallotSize < ballotSize) {
    const padding = new ShortUniqueId({ length: ballotSize - encodedBallotSize });
    encodedBallot += padding();
  }

  const ballotChunks: string[] = [];

  // divide into chunks of 29 bytes, where 1 character === 1 byte
  for (let i = 0; i < ballotSize; i += chunkSize) {
    ballotChunks.push(encodedBallot.substring(i, i + chunkSize));
  }

  return ballotChunks;
}
