import { Buffer } from 'buffer';
import ShortUniqueId from 'short-unique-id';
import { Answers, RANK, SELECT, TEXT } from 'types/configuration';

export function voteEncode(
  answers: Answers,
  ballotSize: number,
  chunksPerBallot: number
): string[] {
  // contains the special string representation of the result
  let encodedBallot = '';

  answers.SelectAnswers.forEach((selectAnswer, id) => {
    encodedBallot += SELECT + ':' + Buffer.from(id).toString('base64') + ':';
    selectAnswer.forEach((answer) => (encodedBallot += answer ? '1,' : '0,'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.RankAnswers.forEach((rankAnswer, id) => {
    encodedBallot += RANK + ':' + Buffer.from(id).toString('base64') + ':';
    const position = Array<number>(rankAnswer.length);
    for (let i = 0; i < rankAnswer.length; i++) {
      position[rankAnswer[i]] = i;
    }
    position.forEach((pos) => (encodedBallot += pos + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.TextAnswers.forEach((textAnswer, id) => {
    encodedBallot += TEXT + ':' + Buffer.from(id).toString('base64') + ':';
    // each answer is first transformed into bytes then encoded in base64
    textAnswer.forEach((answer) => (encodedBallot += Buffer.from(answer).toString('base64') + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  encodedBallot += '\n';

  let encodedBallotSize = Buffer.byteLength(encodedBallot);

  // add padding if necessary until encodedBallot.length == ballotSize
  if (encodedBallotSize < ballotSize) {
    const padding = new ShortUniqueId({ length: ballotSize - encodedBallotSize });
    encodedBallot += padding();
  }

  encodedBallotSize = Buffer.byteLength(encodedBallot);

  const chunkSize = 29;
  const maxEncodedBallotSize = chunkSize * chunksPerBallot;
  const ballotChunks: string[] = [];

  if (encodedBallotSize > maxEncodedBallotSize) {
    throw new Error(
      `actual encoded ballot size ${encodedBallotSize} is bigger than maximum ballot size ${maxEncodedBallotSize}`
    );
  }

  // divide into chunksPerBallot chunks, where 1 character === 1 byte
  for (let i = 0; i < chunksPerBallot; i += 1) {
    const start = i * chunkSize;
    // substring(start, start + chunkSize), if (start + chunkSize) > string.length
    // then (start + chunkSize) is treated as if it was equal to string.length
    ballotChunks.push(encodedBallot.substring(start, start + chunkSize));
  }

  return ballotChunks;
}
